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

// Package controller responsibilities:business worker for each job according to job events
package controller

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	pkgutilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	v1alpha1apis "volcano.sh/volcano/pkg/apis/batch/v1alpha1"
	clientset "volcano.sh/volcano/pkg/client/clientset/versioned"
	samplescheme "volcano.sh/volcano/pkg/client/clientset/versioned/scheme"
	v1alpha1informers "volcano.sh/volcano/pkg/client/informers/externalversions/batch/v1alpha1"
)

// Controller initialize business agent
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// jobclientset is a clientset for volcano job
	jobclientset clientset.Interface

	// component for resource batch/v1alpha1/Job
	jobsSynced  cache.InformerSynced
	jobsIndexer cache.Indexer

	// component for recycle resources
	businessAgent *businessAgent

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	jobclientset clientset.Interface,
	dryRun bool,
	displayStatistic bool,
	podParallelism int,
	cmCheckInterval int,
	cmCheckTimeout int,
	jobInformer v1alpha1informers.JobInformer,
	stopCh <-chan struct{}) *Controller {
	// Create event broadcaster
	// Add ring-controller types to the default Kubernetes Scheme so Events can be
	// logged for ring-controller types.
	pkgutilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(loggerTypeOne).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	businessAgent, err := newBusinessAgent(kubeclientset, recorder, dryRun, displayStatistic, podParallelism,
		cmCheckInterval, cmCheckTimeout, stopCh)
	if err != nil {
		klog.Fatalf("Error creating business agent: %s", err.Error())
	}

	controller := &Controller{
		kubeclientset: kubeclientset,
		jobclientset:  jobclientset,
		jobsSynced:    jobInformer.Informer().HasSynced,
		jobsIndexer:   jobInformer.Informer().GetIndexer(),
		businessAgent: businessAgent,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Jobs"),
		recorder:      recorder,
	}

	klog.V(loggerTypeOne).Info("Setting up event handlers")
	// Set up an event handler for when Job resources change
	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueueJob(obj, EventAdd)
		},
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				controller.enqueueJob(new, EventUpdate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			controller.enqueueJob(obj, EventDelete)
		},
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, monitorPerformance bool, stopCh <-chan struct{}) error {
	defer pkgutilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	defer c.businessAgent.workqueue.ShuttingDown()
	// monitor performance
	if monitorPerformance {
		go startPerformanceMonitorServer()
	}

	// Wait for the caches to be synced before starting workers
	klog.V(loggerTypeFour).Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.jobsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.V(loggerTypeFour).Info("Starting workers")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runMasterWorker, time.Second, stopCh)
	}

	klog.V(loggerTypeFour).Info("Started workers")
	<-stopCh
	klog.V(loggerTypeFour).Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runMasterWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	klog.V(loggerTypeFour).Info("star to get workqueue", c.workqueue.Len())
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
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			pkgutilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Job resource to be synced.
		if err := c.syncHandler(key); err != nil {
			c.workqueue.Forget(obj)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.V(loggerTypeTwo).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		pkgutilruntime.HandleError(err)
		return true
	}

	return true
}

// enqueueJob takes a Job resource and converts
// it into a namespace/name string which is then put onto the work queue. This method
// should *not* be passed resources of any type other than Job.
func (c *Controller) enqueueJob(obj interface{}, eventType string) {
	var key string
	var err error
	if key, err = c.KeyGenerationFunc(obj, eventType); err != nil {
		pkgutilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// KeyGenerationFunc to generate key
func (c *Controller) KeyGenerationFunc(obj interface{}, eventType string) (string, error) {
	metaData, err := meta.Accessor(obj)

	if err != nil {
		return "", fmt.Errorf("object has no meta: %v", err)
	}

	if len(metaData.GetNamespace()) > 0 {
		return metaData.GetNamespace() + "/" + metaData.GetName() + "/" + eventType, nil
	}
	return metaData.GetName() + "/" + eventType, nil
}

// SplitKeyFunc to splite key by format namespace,jobname,eventType
func (c *Controller) SplitKeyFunc(key string) (namespace, name, eventType string, err error) {
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 2:
		// name only, no namespace
		return "", parts[0], parts[1], nil
	case 3:
		// namespace and name
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected key format: %q", key)
}

func (c *Controller) syncHandler(key string) error {
	klog.V(loggerTypeTwo).Infof("syncHandler start, current key is %s", key)

	namespace, name, eventType, err := c.SplitKeyFunc(key)
	if err != nil {
		return fmt.Errorf("failed to split key: %v", err)
	}

	tempObj, exists, err := c.jobsIndexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return fmt.Errorf("failed to get obj from indexer: %s", key)
	}

	switch eventType {
	case EventAdd:
		err := eventAdd(exists, namespace, name, tempObj, c, key)
		if err != nil {
			return err
		}

	case EventDelete:
		if exists {
			// abnormal
			klog.V(loggerTypeTwo).Infof("undefined condition, exist + delete, current key is %s", key)
			return nil
		}
		// delete job worker from businessAgent
		klog.V(loggerTypeTwo).Infof("not exist + delete, current job is %s/%s", namespace, name)
		c.businessAgent.deleteBusinessWorker(namespace, name)

	case EventUpdate:
		// unnecessary to handle
		err := eventUpdate(exists, tempObj, c, namespace, name, key)
		if err != nil {
			return err
		}
	default:
		// abnormal
		klog.V(loggerTypeTwo).Infof("undefined condition, eventType is %s, current key is %s", eventType, key)
	}

	return nil
}

func eventAdd(exists bool, namespace string, name string, tempObj interface{}, c *Controller, key string) error {
	if exists {
		klog.V(loggerTypeTwo).Infof("exist + add, current job is %s/%s", namespace, name)
		// check if job's corresponding configmap is created successfully via volcano controller
		job, ok := tempObj.(*v1alpha1apis.Job)
		if !ok {
			klog.Error("event add => failed, job transform not ok")
		}
		err := c.createBusinessWorker(job)
		if err != nil {
			return err
		}
	}
	// abnormal
	klog.V(loggerTypeTwo).Infof("undefined condition, not exist + add, current key is %s", key)
	return nil
}

func eventUpdate(exists bool, tempObj interface{}, c *Controller, namespace string, name string, key string) error {
	if exists {
		job, ok := tempObj.(*v1alpha1apis.Job)
		if !ok {
			klog.Error("update event -> failed")
		}
		if string(job.Status.State.Phase) == JobRestartPhase {
			c.businessAgent.deleteBusinessWorker(namespace, name)
		} else if !c.businessAgent.isBusinessWorkerExist(namespace, name) {
			// TODO: more restrictions? outdated job update event is not suitable here
			// TODO: although a job update event won't enqueue twice up to now
			// for job update, if create business worker at job restart phase, the version will be incorrect
			err := c.createBusinessWorker(job)
			if err != nil {
				return err
			}
		}
	}
	klog.V(loggerTypeTwo).Infof("undefined condition, not exist + update, current key is %s", key)
	return nil
}

func (c *Controller) createBusinessWorker(job *v1alpha1apis.Job) error {
	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := c.businessAgent.checkConfigmapCreation(job)
	if err != nil {
		return err
	}

	// retrieve configmap data
	var configmapData RankTable
	jobStartString := cm.Data[ConfigmapKey]
	//
	klog.V(loggerTypeFour).Info("jobstarting==>", jobStartString)
	err = json.Unmarshal([]byte(jobStartString), &configmapData)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if configmapData.Status != ConfigmapCompleted && configmapData.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}

	// create a business worker for current job
	err = c.businessAgent.createBusinessWorker(job)
	if err != nil {
		return err
	}

	return nil
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
	error := server.ListenAndServe()
	if error != nil {
		klog.Error(error)
	}
}
