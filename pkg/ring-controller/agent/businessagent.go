/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package agent for run the logic
package agent

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
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

	"hccl-controller/pkg/ring-controller/common"
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
func NewBusinessAgent(kubeClientSet kubernetes.Interface, recorder record.EventRecorder, config *Config,
	stopCh <-chan struct{}) (*BusinessAgent, error) {
	// create pod informer factory
	labelSelector := labels.Set(map[string]string{
		Key910: Val910,
	}).AsSelector().String()
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClientSet,
		time.Second*common.InformerInterval, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
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

	hwlog.RunLog.Info("start informer factory")
	go podInformerFactory.Start(stopCh)
	hwlog.RunLog.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, businessAgent.podInformer.HasSynced); !ok {
		hwlog.RunLog.Errorf("caches sync failed")
		return businessAgent, fmt.Errorf("caches sync failed")
	}

	return businessAgent, businessAgent.run(config.PodParallelism)
}

// enqueuePod to through the monitoring of POD time,
// the corresponding event information is generated and put into the queue of Agent.
func (b *BusinessAgent) enqueuePod(obj interface{}, eventType string) {
	var name string
	var err error
	if name, err = nameGenerationFunc(obj, eventType); err != nil {
		hwlog.RunLog.Errorf("pod key generation error: %v", err)
		return
	}
	b.Workqueue.AddRateLimited(name)
}

func (b *BusinessAgent) run(threadiness int) error {
	hwlog.RunLog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(b.runMasterWorker, time.Second, b.agentSwitch)
	}
	hwlog.RunLog.Info("Started workers")

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
		hwlog.RunLog.Errorf("syncing '%s' failed: failed to get obj from indexer", podKeyInfo)
		return true
	}

	return b.doWorkByWorker(tmpObj, obj, podExist, podKeyInfo)
}

func (b *BusinessAgent) doWorkByWorker(tmpObj, obj interface{}, podExist bool, podKeyInfo *podIdentifier) bool {
	// Lock to safely obtain worker data in the Map
	b.RwMutex.RLock()
	defer b.RwMutex.RUnlock()
	bsnsWorker, workerExist := b.BusinessWorker[podKeyInfo.namespace+"/"+podKeyInfo.jobName]
	hwlog.RunLog.Debugf(" worker : \n %+v", b.BusinessWorker)
	if !workerExist {
		if !podExist {
			b.Workqueue.Forget(obj)
			hwlog.RunLog.Infof("syncing '%s' terminated: current obj is no longer exist",
				podKeyInfo.String())
			return true
		}
		// if someone create a single 910 pod without a job, how to handle?
		hwlog.RunLog.Debugf("syncing '%s' delayed: corresponding job worker may be uninitialized",
			podKeyInfo.String())
		return false
	}
	if podKeyInfo.eventType == EventDelete {
		b.Workqueue.Forget(obj)
		if err := bsnsWorker.handleDeleteEvent(podKeyInfo); err != nil {
			// only logs need to be recorded.
			hwlog.RunLog.Errorf("handleDeleteEvent error, error is %s", err)
		}
		return true
	}
	// if worker exist but pod not exist, try again except delete event
	if !podExist {
		return true
	}
	pod, ok := tmpObj.(*apiCoreV1.Pod)
	if !ok {
		hwlog.RunLog.Error("pod transform failed")
		return true
	}

	// if worker exist && pod exist, need check some special scenarios
	hwlog.RunLog.Debugf("successfully synced '%s'", podKeyInfo)

	forgetQueue, retry := bsnsWorker.doWork(pod, podKeyInfo)
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

func splitWorkerKey(key string) (*podIdentifier, error) {
	parts := strings.Split(key, "/")
	if len(parts) != splitNum {
		return nil, fmt.Errorf("unexpected key format: %q", key)
	}
	podInfo := &podIdentifier{
		namespace: parts[common.Index0],
		name:      parts[common.Index1],
		jobName:   parts[common.Index2],
		eventType: parts[common.Index3],
	}
	return podInfo, nil
}

func preCheck(obj interface{}) (*podIdentifier, bool) {
	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		hwlog.RunLog.Errorf("expected string in WorkerQueue but got %#v", obj)
		return nil, true
	}
	podPathInfo, err := splitWorkerKey(key)
	if err != nil || podPathInfo == nil {
		hwlog.RunLog.Errorf("failed to split key: %v", err)
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
	_, exist := pod.Annotations[PodDeviceKey]
	if !exist {
		hwlog.RunLog.Infof("syncing '%s' delayed: device info is not ready", identifier)
		return false
	}
	return true
}

func containerUsedChip(pod *apiCoreV1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if GetNPUNum(container) > 0 {
			return true
		}
	}

	return false
}

// GetNPUNum get npu npuNum from container:
// 0 presents not use npu;
// -1 presents got invalid npu num;
// other values present use npu;
func GetNPUNum(c apiCoreV1.Container) int32 {
	var qtt resource.Quantity
	var exist bool
	for _, res := range GetResourceList() {
		qtt, exist = c.Resources.Limits[apiCoreV1.ResourceName(res)]
		if !exist {
			continue
		}
		if common.A800MaxChipNum < qtt.Value() || qtt.Value() < 0 {
			return InvalidNPUNum
		}
		return int32(qtt.Value())
	}
	return 0
}

// DeleteWorker : Delete worker(namespace/name) from BusinessWorker map in agent
func DeleteWorker(namespace string, name string, agent *BusinessAgent) {
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()
	hwlog.RunLog.Infof("not exist + delete, current job is %s/%s", namespace, name)
	identifier := namespace + "/" + name
	worker, exist := agent.BusinessWorker[identifier]
	if !exist {
		hwlog.RunLog.Infof("failed to delete business worker for %s/%s, it's not exist", namespace,
			name)
		return
	}

	if agent.Config.DisplayStatistic {
		worker.CloseStatistic()
	}
	delete(agent.BusinessWorker, identifier)
	hwlog.RunLog.Infof("business worker for %s is deleted", identifier)
	return
}
