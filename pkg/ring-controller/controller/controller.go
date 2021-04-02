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

// Package controller responsibilities:business worker for each job according to job events
package controller

import (
	"fmt"
	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/model"
	corev1 "k8s.io/api/core/v1"
	pkgutilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"
	"strings"
	"time"
	clientset "volcano.sh/volcano/pkg/client/clientset/versioned"
	samplescheme "volcano.sh/volcano/pkg/client/clientset/versioned/scheme"
)

// NewController returns a new sample controller
func NewController(kubeclientset kubernetes.Interface, jobclientset clientset.Interface, config *agent.Config,
	informerInfo InformerInfo,
	stopCh <-chan struct{}) *Controller {
	// Create event broadcaster
	// Add ring-controller types to the default Kubernetes Scheme so Events can be
	// logged for ring-controller types.
	pkgutilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(L1).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	agents, err := agent.NewBusinessAgent(kubeclientset, recorder, config, stopCh)
	if err != nil {
		klog.Fatalf("Error creating business agent: %s", err.Error())
	}
	c := &Controller{
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
// is closed, at which point it will shutdown the WorkQueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, monitorPerformance bool, stopCh <-chan struct{}) error {
	defer pkgutilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	defer c.agent.Workqueue.ShuttingDown()
	// monitor performance
	if monitorPerformance {
		go startPerformanceMonitorServer()
	}

	// Wait for the caches to be synced before starting workers
	klog.V(L4).Info("Waiting for informer caches to sync")
	ok := cache.WaitForCacheSync(stopCh, c.jobsSynced, c.deploySynced)
	if !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.V(L4).Info("Starting workers")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runMasterWorker, time.Second, stopCh)
	}

	klog.V(L4).Info("Started workers")
	if stopCh != nil {
		<-stopCh
	}
	klog.V(L4).Info("Shutting down workers")

	return nil
}

// runMasterWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runMasterWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the SyncHandler.
func (c *Controller) processNextWorkItem() bool {
	klog.V(L4).Info("start to get workqueue", c.workqueue.Len())
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var mo model.ResourceEventHandler
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if mo, ok = obj.(model.ResourceEventHandler); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			return fmt.Errorf("expected string in workqueue but got %#v", obj)
		}
		// Run the SyncHandler, passing it the namespace/name string of the
		// Job/Deployment resource to be synced.
		if err := c.SyncHandler(mo); err != nil {
			c.workqueue.Forget(obj)
			return fmt.Errorf("error syncing '%s': %s", mo.GetModelKey(), err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.V(L4).Infof("Successfully synced %+v ", mo)
		return nil
	}(obj)

	if err != nil {
		klog.Errorf("controller processNextWorkItem is failed, err %v", err)
		pkgutilruntime.HandleError(err)
		return true
	}

	return true
}

// enqueueJob takes a Job resource and converts
// it into a namespace/name string which is then put onto the work queue. This method
// should *not* be passed resources of any type other than Job.
func (c *Controller) enqueueJob(obj interface{}, eventType string) {
	models, err := model.Factory(obj, eventType, c.cacheIndexers)
	if err != nil {
		pkgutilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(models)
}

// SyncHandler : to do things from model
func (c *Controller) SyncHandler(model model.ResourceEventHandler) error {
	key := model.GetModelKey()
	klog.V(L2).Infof("SyncHandler start, current key is %v", key)
	namespace, name, eventType, err := splitKeyFunc(key)
	if err != nil {
		return fmt.Errorf("failed to split key: %v", err)
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
		klog.V(L2).Infof("exist + add, current job is %s/%s", namespace, name)
		err := model.EventAdd(c.agent)
		if err != nil {
			return err
		}
	case agent.EventUpdate:
		// unnecessary to handle
		err := model.EventUpdate(c.agent)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}

	return nil
}

func (in *InformerInfo) addEventHandle(controller *Controller) {
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

// splitKeyFunc to splite key by format namespace,jobname,eventType
func splitKeyFunc(key string) (namespace, name, eventType string, err error) {
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 2:
		// name only, no namespace
		return "", parts[0], parts[1], nil
	case 3:
		// namespace and name
		return parts[0], parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("unexpected key format: %q", key)
	}
}

func startPerformanceMonitorServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    net.JoinHostPort("localhost", "6060"),
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		klog.Error(err)
	}
}
