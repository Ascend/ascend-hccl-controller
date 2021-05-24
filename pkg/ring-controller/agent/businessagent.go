/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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

// Package agent for run the logic
package agent

import (
	"fmt"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"strings"
	"time"

	"reflect"
)

// String  to return podIdentifier string style :
// namespace:%s,name:%s,jobName:%s,eventType:%s
func (p *podIdentifier) String() string {
	return fmt.Sprintf("namespace:%s,name:%s,jobName:%s,eventType:%s", p.namespace, p.name, p.jobName, p.eventType)
}

// NewBusinessAgent to create a agent. Agent is a framework, all types of workers can be
// implemented in the form of worker interface in the agent framework run.
// Agent monitors POD events with a specific label and implements the
// combination of tasks through different workers at different times.
var NewBusinessAgent = func(
	kubeClientSet kubernetes.Interface,
	recorder record.EventRecorder,
	config *Config,
	stopCh <-chan struct{}) (*BusinessAgent, error) {

	// create pod informer factory
	labelSelector := labels.Set(map[string]string{
		Key910: Val910,
	}).AsSelector().String()
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClientSet, time.Second*30,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector
		}))

	// each worker share the same init parameters stored here
	businessAgent := &BusinessAgent{
		informerFactory: podInformerFactory,
		podInformer:     podInformerFactory.Core().V1().Pods().Informer(),
		PodsIndexer:     podInformerFactory.Core().V1().Pods().Informer().GetIndexer(),
		Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
			retryMilliSecond*time.Millisecond, threeMinutes*time.Second), "Pods"),
		KubeClientSet:  kubeClientSet,
		BusinessWorker: make(map[string]Worker),
		recorder:       recorder,
		Config:         config,
		agentSwitch:    stopCh,
	}

	// when pod is added, annotation info is ready. No need to listen update event.
	businessAgent.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			businessAgent.enqueuePod(obj, EventAdd)
		},
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				businessAgent.enqueuePod(new, EventUpdate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			businessAgent.enqueuePod(obj, EventDelete)
		},
	})

	klog.V(L1).Info("start informer factory")
	go podInformerFactory.Start(stopCh)
	klog.V(L1).Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, businessAgent.podInformer.HasSynced); !ok {
		klog.Errorf("caches sync failed")
	}

	return businessAgent, businessAgent.run(config.PodParallelism)
}

// enqueuePod to through the monitoring of POD time,
// the corresponding event information is generated and put into the queue of Agent.
func (b *BusinessAgent) enqueuePod(obj interface{}, eventType string) {
	var name string
	var err error
	if name, err = nameGenerationFunc(obj, eventType); err != nil {
		klog.Errorf("pod key generation error: %v", err)
		return
	}
	b.Workqueue.AddRateLimited(name)
}

func (b *BusinessAgent) run(threadiness int) error {
	klog.V(L1).Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(b.runMasterWorker, time.Second, b.agentSwitch)
	}
	klog.V(L1).Info("Started workers")

	return nil
}

func (b *BusinessAgent) runMasterWorker() {
	for b.processNextWorkItem() {
	}
}

func (b *BusinessAgent) processNextWorkItem() bool {
	obj, shutdown := b.Workqueue.Get()

	if shutdown {
		return false
	}

	if !b.doWork(obj) {
		b.Workqueue.AddRateLimited(obj)
	}

	return true
}

// doWork : Each POD time is resolved in detail. If the return value is false, it means that this POD event cannot be
// processed temporarily due to some factors and needs to be put into the queue to continue the next execution.
func (b *BusinessAgent) doWork(obj interface{}) bool {
	// This value is deleted from the queue each time the doWork function is executed.
	defer b.Workqueue.Done(obj)
	// Check the validity of the value in the queue, and if it returns true, discard the value in the queue.
	podKeyInfo, retry := preCheck(obj)
	if retry {
		b.Workqueue.Forget(obj)
		return retry
	}
	// get pod obj from lister
	tmpObj, podExist, err := b.PodsIndexer.GetByKey(podKeyInfo.namespace + "/" + podKeyInfo.name)
	if err != nil {
		b.Workqueue.Forget(obj)
		klog.Errorf("syncing '%s' failed: failed to get obj from indexer", podKeyInfo)
		return true
	}
	// Lock to safely obtain worker data in the Map
	b.RwMutex.RLock()
	defer b.RwMutex.RUnlock()
	bsnsWorker, workerExist := b.BusinessWorker[podKeyInfo.namespace+"/"+podKeyInfo.jobName]
	klog.V(L4).Infof(" worker : \n %+v", b.BusinessWorker)
	if !workerExist {
		if !podExist {
			b.Workqueue.Forget(obj)
			klog.V(L3).Infof("syncing '%s' terminated: current obj is no longer exist",
				podKeyInfo.String())
			return true
		}
		// llTODO: if someone create a single 910 pod without a job, how to handle?
		klog.V(L4).Infof("syncing '%s' delayed: corresponding job worker may be uninitialized",
			podKeyInfo.String())
		return false
	}
	// if worker exist but pod not exist, try again
	if !podExist {
		return true
	}
	pod, ok := tmpObj.(*apiCoreV1.Pod)
	if !ok {
		klog.Error("pod transform failed")
		return true
	}

	// if worker exist && pod exist, need check some special scenarios
	klog.V(L4).Infof("successfully synced '%s'", podKeyInfo)

	forgetQueue, retry := bsnsWorker.doWorker(pod, podKeyInfo)
	if forgetQueue {
		b.Workqueue.Forget(obj)
	}
	return retry
}

// nameGenerationFunc: Generate the objects (Strings) to be put into the queue from POD metadata
func nameGenerationFunc(obj interface{}, eventType string) (string, error) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("object has no meta: %v", err)
	}
	labelMaps := metaData.GetLabels()
	return metaData.GetNamespace() + "/" + metaData.GetName() + "/" + getWorkName(labelMaps) + "/" + eventType, nil
}

func splitWorkerKey(key string) (podInfo *podIdentifier, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != splitNum {
		return nil, fmt.Errorf("unexpected key format: %q", key)
	}
	podInfo = &podIdentifier{
		namespace: parts[0],
		name:      parts[1],
		jobName:   parts[2],
		eventType: parts[3],
	}
	return podInfo, nil
}

func preCheck(obj interface{}) (*podIdentifier, bool) {
	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		klog.Errorf("expected string in WorkerQueue but got %#v", obj)
		return nil, true
	}
	podPathInfo, err := splitWorkerKey(key)
	if err != nil || podPathInfo == nil {
		klog.Errorf("failed to split key: %v", err)
		return nil, true
	}
	return podPathInfo, false
}

func isReferenceJobSameWithBsnsWorker(pod *apiCoreV1.Pod, jobName, bsnsWorkerUID string) bool {
	sameWorker := false
	for _, owner := range pod.OwnerReferences {
		if owner.Name == jobName && string(owner.UID) == bsnsWorkerUID {
			sameWorker = true
			break
		}
	}
	return sameWorker
}

func isPodAnnotationsReady(pod *apiCoreV1.Pod, identifier string) bool {
	useChip := false
	for _, container := range pod.Spec.Containers {
		if GetNPUNum(container) > 0 {
			useChip = true
			break
		}
	}
	if useChip {
		_, exist := pod.Annotations[PodDeviceKey]
		if !exist {
			klog.V(L3).Infof("syncing '%s' delayed: device info is not ready", identifier)
			return false
		}
	}
	return true
}

// GetNPUNum get npu npuNum from container
func GetNPUNum(c apiCoreV1.Container) int32 {
	var qtt resource.Quantity
	var exist bool
	for _, res := range ResourceList {
		qtt, exist = c.Resources.Limits[apiCoreV1.ResourceName(res)]
		if exist && int32(qtt.Value()) > 0 {
			return int32(qtt.Value())
		}
	}
	return 0
}

// DeleteWorker : Delete worker(namespace/name) from BusinessWorker map in agent
func DeleteWorker(namespace string, name string, agent *BusinessAgent) {
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()
	klog.V(L2).Infof("not exist + delete, current job is %s/%s", namespace, name)
	identifier := namespace + "/" + name
	_, exist := agent.BusinessWorker[identifier]
	if !exist {
		klog.V(L3).Infof("failed to delete business worker for %s/%s, it's not exist", namespace,
			name)
		return
	}

	if agent.Config.DisplayStatistic {
		agent.BusinessWorker[identifier].CloseStatistic()
	}
	delete(agent.BusinessWorker, identifier)
	klog.V(L2).Infof("business worker for %s is deleted", identifier)
	return
}
