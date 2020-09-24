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

// Package controller for logic
package controller

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"volcano.sh/volcano/pkg/apis/batch/v1alpha1"

	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"strconv"
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

func newBusinessWorker(kubeclientset kubernetes.Interface, podsIndexer cache.Indexer, recorder record.EventRecorder,
	dryRun bool, job *v1alpha1.Job) *businessWorker {
	var replicasTotal int32
	var groupList []*Group

	for _, taskSpec := range job.Spec.Tasks {
		var deviceTotal int32
		for _, container := range taskSpec.Template.Spec.Containers {
			quantity, exist := container.Resources.Limits[ReeourceName]
			quantityValue := int32(quantity.Value())
			if exist && quantityValue > 0 {
				deviceTotal += quantityValue
			}
		}
		deviceTotal *= taskSpec.Replicas

		var instanceList []*Instance
		group := Group{GroupName: taskSpec.Name, DeviceCount: strconv.FormatInt(int64(deviceTotal), decimal),
			InstanceCount: strconv.FormatInt(int64(taskSpec.Replicas), decimal), InstanceList: instanceList}
		groupList = append(groupList, &group)
		replicasTotal += taskSpec.Replicas
	}

	businessWorker := &businessWorker{
		kubeclientset:        kubeclientset,
		podsIndexer:          podsIndexer,
		recorder:             recorder,
		dryRun:               dryRun,
		statisticSwitch:      make(chan struct{}),
		jobUID:               string(job.UID),
		jobVersion:           job.Status.Version,
		jobCreationTimestamp: job.CreationTimestamp,
		jobNamespace:         job.Namespace,
		jobName:              job.Name,
		configmapName:        fmt.Sprintf("%s-%s", ConfigmapPrefix, job.Name),
		configmapData: RankTable{Status: ConfigmapInitializing, GroupCount: strconv.Itoa(len(job.Spec.Tasks)),
			GroupList: groupList},
		statisticStopped:  false,
		cachedPodNum:      0,
		taskReplicasTotal: replicasTotal,
	}

	return businessWorker
}

func (b *businessWorker) tableConstructionFinished() bool {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()

	return b.cachedPodNum == b.taskReplicasTotal
}

func (b *businessWorker) syncHandler(pod *apiCoreV1.Pod, podExist bool, namespace, name, eventType string) error {
	klog.V(loggerTypeThree).Infof("syncHandler start, current pod is %s/%s, event type is %s", namespace, name,
		eventType)

	// if use 0 chip, end pod sync
	if b.taskReplicasTotal == 0 && b.tableConstructionFinished() {
		klog.V(loggerTypeTwo).Infof("job %s/%s doesn't use d chip, rank table construction is finished",
			b.jobNamespace, b.jobName)
		if err := b.endRankTableConstruction(); err != nil {
			return err
		}
	}

	// dryRun is for test
	if b.dryRun {
		klog.V(loggerTypeThree).Infof("I'am handling %s/%s, event type: %s, exist: %t", namespace, name,
			eventType, podExist)
		return nil
	}

	var err error
	if (eventType == EventAdd) && podExist {
		err := eventAddUpdate(eventType, namespace, name, pod, err, b)
		if err != nil {
			return err
		}
	} else if eventType == EventDelete && !podExist {
		err := eventDelete(namespace, name, b, err)
		if err != nil {
			return err
		}
	}
	klog.V(loggerTypeThree).Infof("undefined condition, pod: %s/%s, event type: %s, exist: %t", namespace,
		name, eventType, podExist)

	return nil
}

func eventDelete(namespace string, name string, b *businessWorker, err error) error {
	klog.V(loggerTypeThree).Infof("not exist + delete, current pod is %s/%s", namespace, name)
	if !b.dryRun {
		// TODO: add task name to key for better forloop efficiency ??
		err = b.removePodInfo(namespace, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func eventAddUpdate(eventType string, namespace string, name string, pod *apiCoreV1.Pod, err error, b *businessWorker) error {
	klog.V(loggerTypeThree).Infof("exist + %s, current pod is %s/%s", eventType, namespace, name)
	// because this annotation is already used to filter pods in previous step (podExist - scenario C)
	// it can be used to identify if pod use chip here
	err = addUpdateEvent(pod, err, b)
	if err != nil {
		return err
	}
	return nil
}

func addUpdateEvent(pod *apiCoreV1.Pod, err error, b *businessWorker) error {
	deviceInfo, exist := pod.Annotations[PodDeviceKey]
	klog.V(loggerTypeThree).Info("deviceId =>", deviceInfo)
	klog.V(loggerTypeFour).Info("isExist ==>", exist)
	if exist {
		err = b.cachePodInfo(pod, deviceInfo)
		if err != nil {
			return err
		}
		return nil
	}

	err = b.cacheZeroChipPodInfo(pod)
	if err != nil {
		return err
	}
	return err
}

// when pod is added, cache current pod info, check if all pods are cached, if true, update configmap
func (b *businessWorker) cachePodInfo(pod *apiCoreV1.Pod, deviceInfo string) error {
	b.cmMu.Lock()
	defer b.cmMu.Unlock()

	for _, group := range b.configmapData.GroupList {
		// find pod's belonging task
		if group.GroupName == pod.Annotations[PodGroupKey] {
			// check if current pod's info is already cached
			for _, instance := range group.InstanceList {
				if instance.PodName == pod.Name {
					klog.V(loggerTypeThree).Infof("ANOMALY: pod %s/%s is already cached", pod.Namespace,
						pod.Name)
					return nil
				}
			}
			// if pod use D chip, cache its info
			var instance Instance
			klog.V(loggerTypeThree).Info("devicedInfo  from pod => %v", deviceInfo)
			err := json.Unmarshal([]byte(deviceInfo), &instance)
			klog.V(loggerTypeThree).Info("instace  from pod => %v", instance)
			if err != nil {
				return fmt.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
			}
			group.InstanceList = append(group.InstanceList, &instance)
			b.modifyStatistics(1)

			// update configmap if finishing caching all pods' info
			if b.tableConstructionFinished() {
				if err := b.endRankTableConstruction(); err != nil {
					return err
				}
			}
			break
		}
	}

	return nil
}

// when pod is added, cache current pod info, check if all pods are cached, if true, update configmap
func (b *businessWorker) cacheZeroChipPodInfo(pod *apiCoreV1.Pod) error {
	b.cmMu.Lock()
	defer b.cmMu.Unlock()

	for _, group := range b.configmapData.GroupList {
		// find pod's belonging task
		if group.GroupName == pod.Annotations[PodGroupKey] {
			// check if current pod's info is already cached
			done := checkPodCache(group, pod)
			if done {
				return nil
			}
			// if pod use D chip, cache its info
			var deviceList []Device
			instance := Instance{PodName: pod.Name, ServerID: "", Devices: deviceList}
			group.InstanceList = append(group.InstanceList, &instance)
			b.modifyStatistics(1)

			// update configmap if finishing caching all pods' info
			errs := updateWithFinish(b)
			if errs != nil {
				return errs
			}
			break
		}
	}

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
			klog.V(loggerTypeThree).Infof("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
			return true
		}
	}
	return false
}

// when pod is deleted, remove pod info from cache, and change configmap's status accordingly
func (b *businessWorker) removePodInfo(namespace string, name string) error {
	b.cmMu.Lock()
	defer b.cmMu.Unlock()
	hasInfoToRemove := false

	for _, group := range b.configmapData.GroupList {
		for idx, instance := range group.InstanceList {
			// current pod's info is already cached, start to remove
			if instance.PodName == name {
				length := len(group.InstanceList)
				group.InstanceList[idx] = group.InstanceList[length-1]
				group.InstanceList = group.InstanceList[:length-1]
				hasInfoToRemove = true
				break
			}
		}
		if hasInfoToRemove {
			break
		}
	}
	if !hasInfoToRemove {
		klog.V(loggerTypeThree).Infof("no data of pod %s/%s can be removed", namespace, name)
		return nil
	}

	klog.V(loggerTypeThree).Infof("start to remove data of pod %s/%s", namespace, name)
	err := b.updateConfigmap()
	if err != nil {
		return err
	}
	b.modifyStatistics(-1)
	klog.V(loggerTypeThree).Infof("data of pod %s/%s is removed", namespace, name)

	return nil
}

func (b *businessWorker) endRankTableConstruction() error {
	b.configmapData.Status = ConfigmapCompleted
	err := b.updateConfigmap()
	if err != nil {
		klog.Error("update configmap failed")
		return err
	}
	klog.V(loggerTypeTwo).Infof("rank table for job %s/%s has finished construction", b.jobNamespace, b.jobName)
	return nil
}

// statistic about how many pods have already cached
func (b *businessWorker) modifyStatistics(diff int32) {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()
	b.cachedPodNum += diff
	klog.V(loggerTypeThree).Infof("rank table build progress for %s/%s: pods need to be cached = %d, "+
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
func (b *businessWorker) statistic() {
	for {
		select {
		case c, ok := <-b.statisticSwitch:
			if !ok {
				klog.Error(c)
			}
			return
		default:
			if b.taskReplicasTotal == b.cachedPodNum {
				klog.V(loggerTypeOne).Infof("rank table build progress for %s/%s is completed",
					b.jobNamespace, b.jobName)
				b.closeStatistic()
			}
			klog.V(loggerTypeOne).Infof("rank table build progress for %s/%s:"+
				" pods need to be cached = %d,pods already cached = %d", b.jobNamespace, b.jobName, b.taskReplicasTotal, b.cachedPodNum)
			time.Sleep(BuildStatInterval)
		}
	}

}
