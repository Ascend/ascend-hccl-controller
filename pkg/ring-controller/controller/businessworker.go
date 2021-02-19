/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
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

// Package controller for logic
package controller

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// controller for each volcano job, list/watch corresponding pods and build configmap (rank table)
type businessWorker struct {
	kubeclientset     kubernetes.Interface
	recorder          record.EventRecorder
	cmMu, statisticMu sync.Mutex
	dryRun            bool
	statisticSwitch   chan struct{}

	podsIndexer cache.Indexer

	// jobVersion: When a job restart, jobVersion is needed to identify if a pod is old
	// with respect to this job
	jobVersion int32
	// jobUID: For an identical job, create it immediately after deletion, new
	// businessWorker will cache old pod info without a identifier to distinguish
	jobUID string
	// jobCreationTimestamp: when pod reference job uid is different with uid of businessWorker
	// creationTimestamp is needed to distinguish cases between: 1. old pod + new worker  OR  2. new pod + old worker
	jobCreationTimestamp metav1.Time
	jobNamespace         string
	jobName              string
	configmapName        string
	configmapData        RankTable

	statisticStopped  bool
	cachedPodNum      int32
	taskReplicasTotal int32
}

func (b *businessWorker) tableConstructionFinished() bool {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()

	return b.cachedPodNum == b.taskReplicasTotal
}

func (b *businessWorker) syncHandler(pod *apiCoreV1.Pod, podExist bool, podInfo *podIdentifier) error {
	klog.V(L3).Infof("syncHandler start, current pod is %s", podInfo)

	// if use 0 chip, end pod sync
	if b.taskReplicasTotal == 0 && b.tableConstructionFinished() {
		klog.V(L2).Infof("job %s/%s doesn't use d chip, rank table construction is finished",
			b.jobNamespace, b.jobName)
		if err := b.endRankTableConstruction(); err != nil {
			return err
		}
		return nil //  need return directly
	}

	// dryRun is for test
	if b.dryRun {
		klog.V(L3).Infof("I'am handling %s, exist: %t", podInfo, podExist)
		return nil
	}

	if (podInfo.eventType == EventAdd) && podExist {
		err := b.handleAddUpdateEvent(podInfo, pod)
		if err != nil {
			return err
		}
	}
	if podInfo.eventType == EventDelete && !podExist {
		err := b.handleDeleteEvent(podInfo)
		if err != nil {
			return err
		}
	}
	klog.V(L3).Infof("undefined condition, pod: %s, exist: %t", podInfo, podExist)
	return nil
}

func (b *businessWorker) handleAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error {
	klog.V(L3).Infof("current addUpdate pod is %s", podInfo)
	// because this annotation is already used to filter pods in previous step (podExist - scenario C)
	// it can be used to identify if pod use chip here
	deviceInfo, exist := pod.Annotations[PodDeviceKey]
	klog.V(L3).Info("deviceId =>", deviceInfo)
	klog.V(L4).Info("isExist ==>", exist)

	b.cmMu.Lock()
	defer b.cmMu.Unlock()

	err := b.configmapData.cachePodInfo(pod, deviceInfo)
	if err != nil {
		return err
	}

	b.modifyStatistics(1)
	// update configmap if finishing caching all pods' info
	errs := updateWithFinish(b)
	if errs != nil {
		return errs
	}

	return nil
}

func (b *businessWorker) handleDeleteEvent(podInfo *podIdentifier) error {
	klog.V(L3).Infof("current handleDeleteEvent pod is %s", podInfo)

	b.cmMu.Lock()
	defer b.cmMu.Unlock()

	err := b.configmapData.removePodInfo(podInfo.namespace, podInfo.name)
	if err != nil {
		return err
	}

	klog.V(L3).Infof("start to remove data of pod %s/%s", podInfo.namespace, podInfo.name)
	err = b.updateConfigmap()
	if err != nil {
		return err
	}
	b.modifyStatistics(-1)
	klog.V(L3).Infof("data of pod %s/%s is removed", podInfo.namespace, podInfo.name)

	return nil
}

func updateWithFinish(b *businessWorker) error {
	if b.tableConstructionFinished() {
		if err := b.endRankTableConstruction(); err != nil {
			return err
		}
	}
	return nil
}

func checkPodCache(group *Group, pod *apiCoreV1.Pod) bool {
	for _, instance := range group.InstanceList {
		if instance.PodName == pod.Name {
			klog.V(L3).Infof("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
			return true
		}
	}
	return false
}

func (b *businessWorker) endRankTableConstruction() error {
	err := b.configmapData.setStatus(ConfigmapCompleted)
	if err != nil {
		klog.Error("fail to set configmap status: %v", err)
		return err
	}
	err = b.updateConfigmap()
	if err != nil {
		klog.Error("update configmap failed")
		return err
	}
	klog.V(L2).Infof("rank table for job %s/%s has finished construction", b.jobNamespace, b.jobName)
	return nil
}

// statistic about how many pods have already cached
func (b *businessWorker) modifyStatistics(diff int32) {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()
	b.cachedPodNum += diff
	klog.V(L3).Infof("rank table build progress for %s/%s: pods need to be cached = %d, "+
		"pods already cached = %d", b.jobNamespace, b.jobName, b.taskReplicasTotal, b.cachedPodNum)
}

// update configmap's data field
func (b *businessWorker) updateConfigmap() error {
	cm, err := b.kubeclientset.CoreV1().ConfigMaps(b.jobNamespace).Get(b.configmapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %v", err)
	}

	label910, exist := (*cm).Labels[Key910]
	if !exist || (exist && label910 != Val910) {
		return fmt.Errorf("invalid configmap label" + label910)
	}

	dataByteArray, err := json.Marshal(b.configmapData)
	if err != nil {
		return fmt.Errorf("marshal configmap data error: %v", err)
	}
	cm.Data[ConfigmapKey] = string(dataByteArray[:])

	if _, err := b.kubeclientset.CoreV1().ConfigMaps(b.jobNamespace).Update(cm); err != nil {
		return fmt.Errorf("failed to update ConfigMap for Job %s/%s: %v", b.jobNamespace, b.jobName, err)
	}

	return nil
}

func (b *businessWorker) closeStatistic() {
	if !b.statisticStopped {
		close(b.statisticSwitch)
		b.statisticStopped = true
	}
}

// no need to add lock here, deviation from true value is acceptable
func (b *businessWorker) statistic(stopTime time.Duration) {
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
					b.jobNamespace, b.jobName)
				b.closeStatistic()
				return
			}
			klog.V(L1).Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", b.jobNamespace, b.jobName, b.taskReplicasTotal, b.cachedPodNum)
			time.Sleep(stopTime)
		}
	}

}
