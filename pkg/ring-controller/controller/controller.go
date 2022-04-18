/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package controller responsibilities:business worker for each job according to job events
package controller

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"huawei.com/npu-exporter/hwlog"
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
	informerInfo InformerInfo,
	stopCh <-chan struct{}) *EventController {
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
		hwlog.RunLog.Fatalf("Error creating business agent: %s", err.Error())
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
	return c
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
	hwlog.RunLog.Debug("get workqueue", c.workqueue.Len())
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
		}
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}

	switch eventType {
	case agent.EventAdd:
		hwlog.RunLog.Infof("exist + add, current job is %s/%s", namespace, name)
		err = model.EventAdd(c.agent)
		if err != nil {
			return err
		}
	case agent.EventUpdate:
		// unnecessary to handle
		err = model.EventUpdate(c.agent)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}

	return nil
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
