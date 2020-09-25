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

// Package controller for run the logic
package controller

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/apis/batch/v1alpha1"

	"hccl-controller/pkg/util/waitcycle"
	"reflect"
	"strconv"
)

// Agent for all businessWorkers, responsibilities:
// * list/watch 910 pods, and assign each pod to corresponding handler
//   (each business worker belongs to a volcano job, and contains a handler for building rank table)
type businessAgent struct {
	informerFactory informers.SharedInformerFactory
	podInformer     cache.SharedIndexInformer
	podsIndexer     cache.Indexer
	kubeClientSet   kubernetes.Interface
	// TODO: use more job info as key will resolve some other todos (e.g. uid)
	// business worker for each volcano job
	businessWorker map[string]*businessWorker

	agentSwitch <-chan struct{}

	rwMu sync.RWMutex

	// event recorder
	recorder record.EventRecorder

	workqueue workqueue.RateLimitingInterface

	// if print only, do not delete anything.
	dryRun bool

	// if display progress of configmap updating
	displayStatistic bool

	// Interval to check job's configmap before building rank table
	cmCheckInterval int

	// Maximum time to check creation of job's configmap
	cmCheckTimeout int
}

type podIdentifier struct {
	namespace string
	name      string
	jobName   string
	eventType string
}

func (p *podIdentifier) String() string {
	return fmt.Sprintf("namespace:%s,name:%s,jobName:%s,eventType:%s", p.namespace, p.name, p.jobName, p.eventType)
}

var newBusinessAgent = func(
	kubeClientSet kubernetes.Interface,
	recorder record.EventRecorder,
	dryRun bool,
	displayStatistic bool,
	podParallelism int,
	cmCheckInterval int,
	cmCheckTimeout int,
	stopCh <-chan struct{}) (*businessAgent, error) {

	// create pod informer factory
	labelSelector := labels.Set(map[string]string{
		Key910: Val910,
	}).AsSelector().String()
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClientSet, time.Second*30,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector
		}))

	// each worker share the same init parameters stored here
	businessAgent := &businessAgent{
		informerFactory: podInformerFactory,
		podInformer:     podInformerFactory.Core().V1().Pods().Informer(),
		podsIndexer:     podInformerFactory.Core().V1().Pods().Informer().GetIndexer(),
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
			retryMilliSecond*time.Millisecond, threeMinutes*time.Second), "Pods"),
		kubeClientSet:    kubeClientSet,
		businessWorker:   make(map[string]*businessWorker),
		recorder:         recorder,
		dryRun:           dryRun,
		displayStatistic: displayStatistic,
		cmCheckInterval:  cmCheckInterval,
		cmCheckTimeout:   cmCheckTimeout,
		agentSwitch:      stopCh,
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

	return businessAgent, businessAgent.run(podParallelism)
}

func (b *businessAgent) enqueuePod(obj interface{}, eventType string) {
	var name string
	var err error
	if name, err = b.nameGenerationFunc(obj, eventType); err != nil {
		klog.Errorf("pod key generation error: %v", err)
		return
	}
	b.workqueue.AddRateLimited(name)
}

func (b *businessAgent) nameGenerationFunc(obj interface{}, eventType string) (string, error) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("object has no meta: %v", err)
	}
	labels := metaData.GetLabels()
	return metaData.GetNamespace() + "/" + metaData.GetName() + "/" + labels[VolcanoJobNameKey] + "/" + eventType, nil
}

func (b *businessAgent) splitKeyFunc(key string) (podInfo *podIdentifier, err error) {
	parts := strings.Split(key, "/")
	if len(parts) == splitNum {
		podInfo := &podIdentifier{
			namespace: parts[0],
			name:      parts[1],
			jobName:   parts[2],
			eventType: parts[3],
		}
		return podInfo, nil
	}
	return nil, fmt.Errorf("unexpected key format: %q", key)
}

func (b *businessAgent) run(threadiness int) error {
	klog.V(L1).Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(b.runMasterWorker, time.Second, b.agentSwitch)
	}
	klog.V(L1).Info("Started workers")

	return nil
}

func (b *businessAgent) runMasterWorker() {
	for b.processNextWorkItem() {
	}
}

func (b *businessAgent) processNextWorkItem() bool {
	obj, shutdown := b.workqueue.Get()

	if shutdown {
		return false
	}

	if !b.doWork(obj) {
		b.workqueue.AddRateLimited(obj)
	}

	return true
}

func (b *businessAgent) doWork(obj interface{}) bool {
	defer b.workqueue.Done(obj)
	podKeyInfo, done := b.preCheck(obj)
	if podKeyInfo == nil {
		return done
	}
	// get pod obj from lister
	tmpObj, podExist, err := b.podsIndexer.GetByKey(podKeyInfo.namespace + "/" + podKeyInfo.name)
	if err != nil {
		b.workqueue.Forget(obj)
		klog.Errorf("syncing '%s' failed: failed to get obj from indexer", podKeyInfo)
		return true
	}

	b.rwMu.RLock()
	defer b.rwMu.RUnlock()
	bsnsWorker, workerExist := b.businessWorker[podKeyInfo.namespace+"/"+podKeyInfo.jobName]
	if !workerExist {
		return b.workerNotExistHandler(podExist, obj, podKeyInfo.String())
	}

	// if worker exist && pod exist, need check some special scenarios
	pod, pass, done := b.convertAndCheckPod(obj, podExist, tmpObj, bsnsWorker, podKeyInfo)
	if !pass {
		return done
	}
	// TODO: pod delete event - new pod not exist + old business worker exist
	// if configmap status of worker struct is completed, no need to sync pod anymore
	pass, done = b.updateConfigMap(obj, pod, podExist, podKeyInfo)
	if !pass {
		return done
	}
	b.workqueue.Forget(obj)
	klog.V(L3).Infof("successfully synced '%s'", podKeyInfo)
	return true
}

func (b *businessAgent) preCheck(obj interface{}) (*podIdentifier, bool) {
	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		b.workqueue.Forget(obj)
		klog.Errorf("expected string in workqueue but got %#v", obj)
		return nil, true
	}
	podPathInfo, err := b.splitKeyFunc(key)
	if err != nil || podPathInfo == nil {
		b.workqueue.Forget(obj)
		klog.Errorf("failed to split key: %v", err)
		return nil, true
	}
	return podPathInfo, false
}

func (b *businessAgent) convertAndCheckPod(obj interface{}, podExist bool, tmpObj interface{}, bsnsWorker *businessWorker, podInfo *podIdentifier) (newPod *v1.Pod, isPass bool, isOver bool) {
	var pod *v1.Pod
	if !podExist {
		return pod, false, true
	}
	var ok bool
	pod, ok = tmpObj.(*v1.Pod)
	if !ok {
		klog.Error("pod transform failed")
		return nil, false, true
	}
	done, pass := b.checkPodCondition(pod, bsnsWorker, obj, podInfo)
	if !pass {
		return pod, false, done
	}
	return pod, true, false

}

func (b *businessAgent) updateConfigMap(obj interface{}, pod *v1.Pod, podExist bool, podInfo *podIdentifier) (pass, isOver bool) {
	if b.businessWorker[podInfo.namespace+"/"+podInfo.jobName].configmapData.Status == ConfigmapCompleted {
		b.workqueue.Forget(obj)
		klog.V(L3).Infof("syncing '%s' terminated: corresponding rank table is completed",
			podInfo)
		return false, true
	}
	// start to sync current pod
	if err := b.businessWorker[podInfo.namespace+"/"+podInfo.jobName].syncHandler(pod, podExist, podInfo); err != nil {
		b.workqueue.Forget(obj)
		klog.Errorf("error syncing '%s': %s", podInfo, err.Error())
		return false, true
	}

	return true, false
}

func (b *businessAgent) checkPodCondition(pod *v1.Pod, bsnsWorker *businessWorker, obj interface{}, podInfo *podIdentifier) (isOver, pass bool) {

	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if !isReferenceJobSameWithBsnsWorker(pod, podInfo.jobName, bsnsWorker.jobUID) {
		if pod.CreationTimestamp.Before(&bsnsWorker.jobCreationTimestamp) {
			// old pod + new worker
			b.workqueue.Forget(obj)
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
		b.workqueue.Forget(obj)
		klog.Errorf("syncing '%s' failed, parse pod annotation error: %v", podInfo, err)
		return true, false
	}
	version32 := int32(version64)
	// job restart action will increase job version number
	if version32 < bsnsWorker.jobVersion {
		b.workqueue.Forget(obj)
		klog.V(L3).Infof("syncing '%s' terminated: corresponding job worker "+
			"is no longer exist (basis: job version number)", podInfo)
		return true, false
	} else if version32 > bsnsWorker.jobVersion {
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
	return false, true
}

func (b *businessAgent) workerNotExistHandler(podExist bool, obj interface{}, key string) bool {
	if !podExist {
		b.workqueue.Forget(obj)
		klog.V(L3).Infof("syncing '%s' terminated: current obj is no longer exist",
			key)
		return true
	}
	// llTODO: if someone create a single 910 pod without a job, how to handle?
	klog.V(L3).Infof("syncing '%s' delayed: corresponding job worker may be uninitialized",
		key)
	return false
}

func isReferenceJobSameWithBsnsWorker(pod *v1.Pod, jobName, bsnsWorkerUID string) bool {
	sameWorker := false
	for _, owner := range pod.OwnerReferences {
		if owner.Name == jobName && string(owner.UID) == bsnsWorkerUID {
			sameWorker = true
			break
		}
	}
	return sameWorker
}

func isPodAnnotationsReady(pod *v1.Pod, identifier string) bool {
	useChip := false
	for _, container := range pod.Spec.Containers {
		quantity, exist := container.Resources.Limits[ResourceName]
		if exist && int(quantity.Value()) > 0 {
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

// CheckConfigmapCreation check configmap
func (b *businessAgent) CheckConfigmapCreation(job *v1alpha1.Job) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	err := waitcycle.Wait(time.Duration(b.cmCheckTimeout)*time.Second, func() (bool, error) {
		var errTmp error
		cm, errTmp = b.kubeClientSet.CoreV1().ConfigMaps(job.Namespace).Get(fmt.Sprintf("%s-%s",
			ConfigmapPrefix, job.Name), metav1.GetOptions{})
		if errTmp != nil {
			if errors.IsNotFound(errTmp) {
				return false, nil
			}
			return true, fmt.Errorf("get configmap error: %v", errTmp)
		}
		return true, nil
	}, time.Duration(b.cmCheckInterval)*time.Second)

	if err != nil {
		return nil, fmt.Errorf("failed to get configmap for job %s/%s: %v", job.Namespace, job.Name, err)
	}
	label910, exist := (*cm).Labels[Key910]
	if !exist || (exist && label910 != Val910) {
		return nil, fmt.Errorf("invalid configmap label" + label910)
	}

	return cm, nil
}

// CreateBusinessWorker create worker
func (b *businessAgent) CreateBusinessWorker(job *v1alpha1.Job) error {
	b.rwMu.Lock()
	defer b.rwMu.Unlock()

	klog.V(L2).Infof("create business worker for %s/%s", job.Namespace, job.Name)

	_, exist := b.businessWorker[job.Namespace+"/"+job.Name]
	if exist {
		klog.V(L2).Infof("business worker for %s/%s is already existed", job.Namespace, job.Name)
		return nil
	}

	// initialize business worker for current job
	businessWorker := newBusinessWorker(b.kubeClientSet, b.podsIndexer, b.recorder, b.dryRun, job)

	// start to report rank table build statistic for current job
	if b.displayStatistic {
		go businessWorker.statistic(BuildStatInterval)
	}

	// save current business worker
	b.businessWorker[job.Namespace+"/"+job.Name] = businessWorker

	klog.V(L2).Infof("create business worker for %s/%s success, %d pods need to be cached",
		job.Namespace, job.Name, b.businessWorker[job.Namespace+"/"+job.Name].taskReplicasTotal)

	return nil
}

// DeleteBusinessWorker delete businessworker
func (b *businessAgent) DeleteBusinessWorker(namespace string, name string) error {
	b.rwMu.Lock()
	defer b.rwMu.Unlock()

	identifier := namespace + "/" + name
	_, exist := b.businessWorker[identifier]
	if !exist {
		klog.V(L2).Infof("failed to delete business worker for %s/%s, it's not exist", namespace,
			name)
		return nil
	}

	if b.displayStatistic {
		b.businessWorker[identifier].closeStatistic()
	}
	delete(b.businessWorker, identifier)
	klog.V(L2).Infof("business worker for %s/%s is deleted", namespace, name)

	return nil
}

// IsBusinessWorkerExist check worker if exist
func (b *businessAgent) IsBusinessWorkerExist(namespace string, name string) bool {
	b.rwMu.Lock()
	defer b.rwMu.Unlock()
	_, exist := b.businessWorker[namespace+"/"+name]
	return exist
}
