/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */
package controller

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	v1alpha1apis "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	vofake "volcano.sh/apis/pkg/client/clientset/versioned/fake"
	informers "volcano.sh/apis/pkg/client/informers/externalversions"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/model"
	_ "hccl-controller/pkg/testtool"
)

// TestControllerRun test Controller Run
func TestControllerRun(t *testing.T) {
	Convey("controller Controller_Run", t, func() {
		ctr := newFakeController()
		Convey("err != nil when cache not exist ", func() {
			patches := ApplyFunc(cache.WaitForCacheSync, func(_ <-chan struct{}, _ ...cache.InformerSynced) bool {
				return false
			})
			defer patches.Reset()
			err := ctr.Run(1, nil)
			So(err, ShouldNotEqual, nil)
		})

		Convey("err == nil when cache exist ", func() {
			patches := ApplyFunc(cache.WaitForCacheSync, func(_ <-chan struct{}, _ ...cache.InformerSynced) bool {
				return true
			})
			defer patches.Reset()
			err := ctr.Run(1, nil)
			So(err, ShouldEqual, nil)
		})
	})
}

// TestProcessNextWorkItem test ProcessNextWorkItem
func TestProcessNextWorkItem(t *testing.T) {
	Convey("controller ProcessNextWorkItem", t, func() {
		ctr := newFakeController()
		Convey("res == true when process  ", func() {
			obj := &v1alpha1apis.Job{metav1.TypeMeta{}, metav1.ObjectMeta{Name: "test1", GenerateName: "",
				Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "", Generation: 0,
				CreationTimestamp: metav1.Now(), DeletionTimestamp: nil, DeletionGracePeriodSeconds: nil, Labels: nil,
				Annotations: nil, OwnerReferences: nil, Finalizers: nil, ClusterName: "", ManagedFields: nil},
				v1alpha1apis.JobSpec{}, v1alpha1apis.JobStatus{}}
			ctr.enqueueJob(obj, agent.EventAdd)
			patches := ApplyMethod(reflect.TypeOf(ctr), "SyncHandler", func(_ *Controller,
				m model.ResourceEventHandler) error {
				return fmt.Errorf("undefined condition, things is %s", m.GetModelKey())
			})
			defer patches.Reset()
			res := ctr.processNextWork()
			So(res, ShouldEqual, true)
			So(ctr.workqueue.Len(), ShouldEqual, 0)
		})

		Convey("err != nil when cache not exist ", func() {
			obj := &v1alpha1apis.Job{metav1.TypeMeta{}, metav1.ObjectMeta{Name: "test1", GenerateName: "",
				Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "", Generation: 0,
				CreationTimestamp: metav1.Now(), DeletionTimestamp: nil, DeletionGracePeriodSeconds: nil, Labels: nil,
				Annotations: nil, OwnerReferences: nil, Finalizers: nil, ClusterName: "", ManagedFields: nil},
				v1alpha1apis.JobSpec{}, v1alpha1apis.JobStatus{}}
			ctr.enqueueJob(obj, agent.EventAdd)
			patches := ApplyMethod(reflect.TypeOf(ctr), "SyncHandler", func(_ *Controller,
				m model.ResourceEventHandler) error {
				return nil
			})
			defer patches.Reset()
			res := ctr.processNextWork()
			So(res, ShouldEqual, true)
			So(ctr.workqueue.Len(), ShouldEqual, 0)
		})
	})
}

// TestControllerSyncHandler test Controller SyncHandler
func TestControllerSyncHandler(t *testing.T) {
	Convey("controller Controller_SyncHandler", t, func() {
		ctr := newFakeController()
		Convey("err != nil when splitKeyFunc return err  ", func() {
			obj := &v1alpha1apis.Job{metav1.TypeMeta{}, metav1.ObjectMeta{Name: "test", GenerateName: "",
				Namespace: "namespace", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "", Generation: 0,
				CreationTimestamp: metav1.Now(), DeletionTimestamp: nil, DeletionGracePeriodSeconds: nil, Labels: nil,
				Annotations: nil, OwnerReferences: nil, Finalizers: nil, ClusterName: "", ManagedFields: nil},
				v1alpha1apis.JobSpec{}, v1alpha1apis.JobStatus{}}
			rs, _ := model.Factory(obj, agent.EventAdd, ctr.cacheIndexers)
			patches := ApplyFunc(splitKeyFunc, func(_ string) (namespace, name, eventType string, err error) {
				return "", "", "", fmt.Errorf("undefined condition")
			})
			defer patches.Reset()
			err := ctr.SyncHandler(rs)
			So(err, ShouldNotEqual, nil)
		})
		Convey("err != nil when index getByKey return err  ", func() {
			obj := &v1alpha1apis.Job{metav1.TypeMeta{}, metav1.ObjectMeta{Name: "test", GenerateName: "",
				Namespace: "namespace", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "", Generation: 0,
				CreationTimestamp: metav1.Now(), DeletionTimestamp: nil, DeletionGracePeriodSeconds: nil, Labels: nil,
				Annotations: nil, OwnerReferences: nil, Finalizers: nil, ClusterName: "", ManagedFields: nil},
				v1alpha1apis.JobSpec{}, v1alpha1apis.JobStatus{}}
			rs, _ := model.Factory(obj, agent.EventAdd, ctr.cacheIndexers)
			rs.GetCacheIndex().Add(obj)
			patches := ApplyMethod(reflect.TypeOf(rs), "EventAdd", func(_ *model.VCJobModel,
				_ *agent.BusinessAgent) error {
				return nil
			})
			defer patches.Reset()
			err := ctr.SyncHandler(rs)
			So(err, ShouldEqual, nil)
		})
	})
}

func newFakeController() *Controller {
	config := newTestConfig()
	kube := fake.NewSimpleClientset()
	volcano := vofake.NewSimpleClientset()
	jobInformerFactory := informers.NewSharedInformerFactoryWithOptions(volcano, time.Second*30,
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			return
		}))
	deploymentFactory := cinformers.NewSharedInformerFactoryWithOptions(kube, time.Second*30,
		cinformers.WithTweakListOptions(func(options *v1.ListOptions) {
			return
		}))
	jobInformer := jobInformerFactory.Batch().V1alpha1().Jobs()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.VCJobType] = jobInformer.Informer().GetIndexer()
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()
	return NewController(kube, volcano, config, InformerInfo{JobInformer: jobInformer,
		DeployInformer: deploymentInformer, CacheIndexers: cacheIndexer}, make(chan struct{}))
}

func newTestConfig() *agent.Config {
	const (
		PodParalle  = 1
		CmCheckIn   = 3
		CmCheckTout = 10
	)
	return &agent.Config{
		DryRun:           false,
		DisplayStatistic: false,
		PodParallelism:   PodParalle,
		CmCheckInterval:  CmCheckIn,
		CmCheckTimeout:   CmCheckTout,
	}
}
