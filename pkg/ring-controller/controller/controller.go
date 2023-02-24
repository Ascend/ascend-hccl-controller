/* Copyright(C) 2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package controller responsibilities:business worker for each job according to job events
package controller

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"huawei.com/npu-exporter/v3/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"volcano.sh/apis/pkg/client/clientset/versioned"
	samplescheme "volcano.sh/apis/pkg/client/clientset/versioned/scheme"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	"hccl-controller/pkg/ring-controller/model"
)

// NewEventController returns a new sample controller
func NewEventController(kubeclientset kubernetes.Interface, jobclientset versioned.Interface, config *agent.Config,
	informerInfo InformerInfo, stopCh <-chan struct{}) (*EventController, error) {
	// Create event broadcaster
	// Add ring-controller types to the default Kubernetes Scheme so Events can be
	// logged for ring-controller types.
	runtime.Must(samplescheme.AddToScheme(scheme.Scheme))
	hwlog.RunLog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(hwlog.RunLog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	agents, err := agent.NewBusinessAgent(kubeclientset, recorder, config, stopCh)
	if err != nil {
		return nil, fmt.Errorf("error creating business agent: %s", err.Error())
	}
	c := &EventController{
		kubeclientset: kubeclientset,
		jobclientset:  jobclientset,
		jobsSynced:    informerInfo.JobInformer.Informer().HasSynced,
		deploySynced:  informerInfo.DeployInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "model"),
		recorder:      recorder,
		agent:         agents,
		cacheIndexers: informerInfo.CacheIndexers,
	}
	informerInfo.addEventHandle(c)
	return c, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
func (c *EventController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()
	defer c.agent.Workqueue.ShuttingDown()

	// Wait for the caches to be synced before starting workers
	hwlog.RunLog.Debug("Waiting for informer caches to sync")
	ok := cache.WaitForCacheSync(stopCh, c.jobsSynced, c.deploySynced)
	if !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	hwlog.RunLog.Debug("Starting workers")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runMaster, time.Second, stopCh)
	}

	hwlog.RunLog.Debug("Started")
	if stopCh != nil {
		<-stopCh
	}
	hwlog.RunLog.Debug("Shutting down")
	return nil
}

func (c *EventController) runMaster() {
	for c.processNextWork() {
	}
}

func (c *EventController) processNextWork() bool {
	hwlog.RunLog.Debug("get workqueue-", c.workqueue.Len())
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {

		defer c.workqueue.Done(obj)
		var mo model.ResourceEventHandler
		var ok bool
		if mo, ok = obj.(model.ResourceEventHandler); !ok {
			c.workqueue.Forget(obj)
			return fmt.Errorf("expected ResourceEventHandler in workqueue but got %#v", obj)
		}

		if err := c.SyncHandler(mo); err != nil {
			c.workqueue.Forget(obj)
			return fmt.Errorf("error to syncing '%s': %s", mo.GetModelKey(), err.Error())
		}

		c.workqueue.Forget(obj)
		hwlog.RunLog.Debugf("Synced Successfully %+v ", mo)
		return nil
	}(obj)

	if err != nil {
		hwlog.RunLog.Errorf("processNextWork controller, err %v", err)
		runtime.HandleError(err)
		return true
	}

	return true
}

// enqueueJob takes a Job resource and converts
// it into a namespace/name string which is then put onto the work queue. This method
// should *not* be passed resources of any type other than Job.
func (c *EventController) enqueueJob(obj interface{}, eventType string) {
	models, err := model.Factory(obj, eventType, c.cacheIndexers)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(models)
}

// SyncHandler : to do things from model
func (c *EventController) SyncHandler(model model.ResourceEventHandler) error {
	key := model.GetModelKey()
	hwlog.RunLog.Infof("SyncHandler start, current key is %v", key)

	var namespace, name, eventType string
	parts := strings.Split(key, "/")
	switch len(parts) {
	case common.Index2:
		// name only, no namespace
		namespace = ""
		name = parts[common.Index0]
		eventType = parts[common.Index1]
	case common.Index3:
		// namespace and name
		namespace = parts[common.Index0]
		name = parts[common.Index1]
		eventType = parts[common.Index2]
	default:
		return fmt.Errorf("failed to split key, unexpected key format: %q", key)
	}

	_, exists, err := model.GetCacheIndex().GetByKey(namespace + "/" + name)
	if err != nil {
		return fmt.Errorf("failed to get obj from indexer: %s", key)
	}
	if !exists {
		if eventType == agent.EventDelete {
			agent.DeleteWorker(namespace, name, c.agent)
			hwlog.RunLog.Infof("not exist + delete, eventType is %s, current key is %s", eventType, key)
			return nil
		}
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}

	switch eventType {
	case agent.EventAdd:
		hwlog.RunLog.Infof("exist + add, current job is %s/%s", namespace, name)
		return model.EventAdd(c.agent)
	case agent.EventUpdate:
		// unnecessary to handle
		return model.EventUpdate(c.agent)
	default:
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}
}

func (in *InformerInfo) addEventHandle(controller *EventController) {
	eventHandlerFunc := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueueJob(obj, agent.EventAdd)
		},
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				controller.enqueueJob(new, agent.EventUpdate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			controller.enqueueJob(obj, agent.EventDelete)
		},
	}
	in.JobInformer.Informer().AddEventHandler(eventHandlerFunc)
	in.DeployInformer.Informer().AddEventHandler(eventHandlerFunc)
}
