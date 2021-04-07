/*
 * Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package agent for logic
package agent

import (
	"encoding/json"
	"fmt"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"strconv"
	"time"
)

// Worker :The main function of Worker is to get the information of NPU from the generated POD,
// and then assemble it into a complete HCCL.JSON file.
type Worker interface {
	doWorker(pod *apiCoreV1.Pod, podInfo *podIdentifier) (forgetQueue, retry bool)
	Statistic(stopTime time.Duration)
	WorkerCommon
}

// NewVCJobWorker : Generates a Worker that handles the VCJob type
func NewVCJobWorker(agent *BusinessAgent, job JobInfo, ranktable v1.RankTabler, replicasTotal int32) *VCJobWorker {
	jobWorker := &VCJobWorker{WorkerInfo: WorkerInfo{kubeclientset: agent.KubeClientSet, podsIndexer: agent.PodsIndexer,
		recorder: agent.recorder, dryRun: agent.dryRun, statisticSwitch: make(chan struct{}),
		configmapName: fmt.Sprintf("%s-%s", ConfigmapPrefix, job.JobName),
		configmapData: ranktable, statisticStopped: false, cachedPodNum: 0, taskReplicasTotal: replicasTotal,
		rankMap: make(map[string]int, 1)}, JobInfo: job}
	return jobWorker
}

func (b *VCJobWorker) doWorker(pod *apiCoreV1.Pod, podInfo *podIdentifier) (forgetQueue, retry bool) {
	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if !isReferenceJobSameWithBsnsWorker(pod, podInfo.jobName, b.JobUID) {
		if pod.CreationTimestamp.Before(&b.JobCreationTimestamp) {
			// old pod + new worker
			klog.V(L3).Infof("syncing '%s' terminated: corresponding job worker is no "+
				"longer exist (basis: job uid + creationTimestamp)", podInfo)
			return true, false
		}
		// new pod + old worker
		klog.V(L3).Infof("syncing '%s' delayed: corresponding job worker is "+
			"uninitialized (basis: job uid + creationTimestamp)", podInfo)
		return false, false

	}
	// scenario check B: job set restart policy, delete pod
	// check basis: job version
	version64, err := strconv.ParseInt(pod.Annotations[PodJobVersion], 10, 32)
	if err != nil {
		klog.Errorf("syncing '%s' failed, parse pod annotation error: %v", podInfo, err)
		return true, false
	}
	version32 := int32(version64)
	// job restart action will increase job version number
	if version32 < b.JobVersion {

		klog.V(L3).Infof("syncing '%s' terminated: corresponding job worker "+
			"is no longer exist (basis: job version number)", podInfo)
		return true, false
	}
	if version32 > b.JobVersion {
		klog.V(L3).Infof("syncing '%s' delayed: corresponding job worker "+
			"is uninitialized (basis: job version number)", podInfo)
		return false, false
	}
	// scenario check C: if current pod use chip, its' device info may not be ready
	// check basis: limits + annotations
	if (podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate) && !isPodAnnotationsReady(pod,
		podInfo.String()) {
		return false, false
	}
	if configmapComplete :=
		b.configmapData.GetStatus() == ConfigmapCompleted; configmapComplete {
		klog.V(L3).Infof("syncing '%s' terminated: corresponding rank table is completed",
			podInfo)
		return true, true
	}

	// start to sync current pod
	if err := b.syncHandler(pod, podInfo); err != nil {
		klog.Errorf("error syncing '%s': %s", podInfo, err.Error())
		return true, true
	}
	return true, true
}

// Statistic : Determine whether CM has been built, process the build completion or change the goroutine exit signal.
// No need to add lock here, deviation from true value is acceptable
func (b *VCJobWorker) Statistic(stopTime time.Duration) {
	for {
		select {
		case c, ok := <-b.statisticSwitch:
			if !ok {
				klog.Error(c)
			}
			return
		default:
			if b.taskReplicasTotal == b.cachedPodNum {
				klog.V(L1).Infof("rank table build progress for %s/%s is completed",
					b.JobNamespace, b.JobName)
				b.CloseStatistic()
				return
			}
			klog.V(L2).Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", b.JobNamespace, b.JobName, b.taskReplicasTotal, b.cachedPodNum)
			time.Sleep(stopTime)
		}
	}
}

// WorkerCommon : The common methods of Worker, these methods have a certain degree of fixedness,
// if the new Worker type does not apply to these methods, they can be overwritten.
type WorkerCommon interface {
	handleAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error
	handleDeleteEvent(podInfo *podIdentifier) error
	tableConstructionFinished() bool
	endRankTableConstruction(string) error
	modifyStatistics(diff int32)
	// CloseStatistic : to close statisticSwitch chan
	CloseStatistic()
	syncHandler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error
}

func (b *WorkerInfo) syncHandler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error {
	klog.V(L3).Infof("syncHandler start, current pod is %s", podInfo)

	// if use 0 chip, end pod sync
	if b.taskReplicasTotal == 0 && b.tableConstructionFinished() {
		klog.V(L2).Infof("job %s/%s doesn't use d chip, rank table construction is finished",
			podInfo.namespace, podInfo.jobName)
		if err := b.endRankTableConstruction(pod.Namespace); err != nil {
			return err
		}
		klog.V(L2).Infof("rank table for job %s/%s has finished construction", podInfo.namespace, podInfo.jobName)
		return nil //  need return directly
	}

	// dryRun is for test
	if b.dryRun {
		klog.V(L3).Infof("I'am handling %s", podInfo)
		return nil
	}

	if podInfo.eventType == EventAdd {
		return b.handleAddUpdateEvent(podInfo, pod)

	}
	if podInfo.eventType == EventDelete {
		return b.handleDeleteEvent(podInfo)

	}
	klog.V(L3).Infof("undefined condition, pod: %s", podInfo)
	return nil
}

func (b *WorkerInfo) tableConstructionFinished() bool {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()

	return b.cachedPodNum == b.taskReplicasTotal
}
func (b *WorkerInfo) handleAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error {
	klog.V(L4).Infof("current addUpdate pod is %s", podInfo)
	// because this annotation is already used to filter pods in previous step (podExist - scenario C)
	// it can be used to identify if pod use chip here
	deviceInfo, exist := pod.Annotations[PodDeviceKey]
	klog.V(L3).Infof("deviceId => %s", deviceInfo)
	klog.V(L4).Infof("isExist ==> %t", exist)

	b.cmMu.Lock()
	defer b.cmMu.Unlock()
	b.rankMap[podInfo.namespace+"/"+podInfo.name] = b.rankIndex
	err := b.configmapData.CachePodInfo(pod, deviceInfo, &b.rankIndex)
	if err != nil {
		return err
	}

	b.modifyStatistics(1)
	klog.V(L3).Infof("rank table build progress for %s/%s: pods need to be cached = %d, "+
		"pods already cached = %d", podInfo.namespace, podInfo.jobName, b.taskReplicasTotal, b.cachedPodNum)
	// update configmap if finishing caching all pods' info
	errs := updateWithFinish(b, podInfo.namespace)
	if errs != nil {
		return errs
	}

	return nil
}

func (b *WorkerInfo) handleDeleteEvent(podInfo *podIdentifier) error {
	klog.V(L3).Infof("current handleDeleteEvent pod is %s", podInfo)

	b.cmMu.Lock()
	defer b.cmMu.Unlock()
	err := b.configmapData.RemovePodInfo(podInfo.namespace, podInfo.name)
	if err != nil {
		return err
	}

	klog.V(L3).Infof("start to remove data of pod %s/%s", podInfo.namespace, podInfo.name)
	err = updateConfigMap(b, podInfo.namespace)
	if err != nil {
		return err
	}
	b.modifyStatistics(-1)
	klog.V(L3).Infof("data of pod %s/%s is removed", podInfo.namespace, podInfo.name)

	return nil
}

func (b *WorkerInfo) endRankTableConstruction(namespace string) error {
	err := b.configmapData.SetStatus(ConfigmapCompleted)
	if err != nil {
		klog.Errorf("fail to set configmap status: %s", err)
		return err
	}
	err = updateConfigMap(b, namespace)
	if err != nil {
		klog.Error("update configmap failed")
		return err
	}
	return nil
}

// modifyStatistics statistic about how many pods have already cached
func (b *WorkerInfo) modifyStatistics(diff int32) {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()
	b.cachedPodNum += diff

}

// CloseStatistic : to close statisticSwitch chan
func (b *WorkerInfo) CloseStatistic() {
	if !b.statisticStopped {
		close(b.statisticSwitch)
		b.statisticStopped = true
	}
}

func updateWithFinish(b *WorkerInfo, namespace string) error {
	if b.tableConstructionFinished() {
		if err := b.endRankTableConstruction(namespace); err != nil {
			return err
		}
	}
	return nil
}

func getWorkName(labels map[string]string) string {
	if label, ok := labels[VolcanoJobNameKey]; ok {
		return label
	}
	return labels[DeploymentNameKey]
}

func updateConfigMap(w *WorkerInfo, namespace string) error {
	cm, err := w.kubeclientset.CoreV1().ConfigMaps(namespace).Get(w.configmapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %v", err)
	}

	label910, exist := (*cm).Labels[Key910]
	if !exist || (exist && label910 != Val910) {
		return fmt.Errorf("invalid configmap label" + label910)
	}

	dataByteArray, err := json.Marshal(w.configmapData)
	if err != nil {
		return fmt.Errorf("marshal configmap data error: %v", err)
	}
	cm.Data[ConfigmapKey] = string(dataByteArray[:])

	if _, err := w.kubeclientset.CoreV1().ConfigMaps(namespace).Update(cm); err != nil {
		return fmt.Errorf("failed to update ConfigMap for Job %v", err)
	}
	return nil
}
