/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
package controller

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	vofake "volcano.sh/apis/pkg/client/clientset/versioned/fake"
	"volcano.sh/apis/pkg/client/informers/externalversions"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	"hccl-controller/pkg/ring-controller/model"
	_ "hccl-controller/pkg/testtool"
)

// TestControllerRun test Controller Run
func TestControllerRun(t *testing.T) {
	convey.Convey("controller Controller_Run", t, func() {
		ctr := newFakeController()
		convey.Convey("err != nil when cache not exist ", func() {
			patches := gomonkey.ApplyFunc(cache.WaitForCacheSync, func(_ <-chan struct{}, _ ...cache.InformerSynced) bool {
				return false
			})
			defer patches.Reset()
			err := ctr.Run(1, nil)
			convey.So(err, convey.ShouldNotEqual, nil)
		})

		convey.Convey("err == nil when cache exist ", func() {
			patches := gomonkey.ApplyFunc(cache.WaitForCacheSync, func(_ <-chan struct{}, _ ...cache.InformerSynced) bool {
				return true
			})
			defer patches.Reset()
			err := ctr.Run(1, nil)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

// TestProcessNextWorkItem test ProcessNextWorkItem
func TestProcessNextWorkItem(t *testing.T) {
	convey.Convey("controller ProcessNextWorkItem", t, func() {
		ctr := newFakeController()
		convey.Convey("res == true when process  ", func() {
			obj := &v1alpha1.Job{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test1",
				GenerateName: "", Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "",
				Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ManagedFields: nil}, Spec: v1alpha1.JobSpec{},
				Status: v1alpha1.JobStatus{}}
			ctr.enqueueJob(obj, agent.EventAdd)
			patches := gomonkey.ApplyMethod(reflect.TypeOf(ctr), "SyncHandler", func(_ *EventController,
				m model.ResourceEventHandler) error {
				return fmt.Errorf("undefined condition, things is %s", m.GetModelKey())
			})
			defer patches.Reset()
			res := ctr.processNextWork()
			convey.So(res, convey.ShouldEqual, true)
			convey.So(ctr.workqueue.Len(), convey.ShouldEqual, 0)
		})

		convey.Convey("err != nil when cache not exist ", func() {
			obj := &v1alpha1.Job{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test1",
				GenerateName: "", Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "",
				Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ManagedFields: nil}, Spec: v1alpha1.JobSpec{},
				Status: v1alpha1.JobStatus{}}
			ctr.enqueueJob(obj, agent.EventAdd)
			patches := gomonkey.ApplyMethod(reflect.TypeOf(ctr), "SyncHandler", func(_ *EventController,
				m model.ResourceEventHandler) error {
				return nil
			})
			defer patches.Reset()
			res := ctr.processNextWork()
			convey.So(res, convey.ShouldEqual, true)
			convey.So(ctr.workqueue.Len(), convey.ShouldEqual, 0)
		})
	})
}

// TestControllerSyncHandler test Controller SyncHandler
func TestControllerSyncHandler(t *testing.T) {
	convey.Convey("controller Controller_SyncHandler", t, func() {
		ctr := newFakeController()
		convey.Convey("err != nil when splitKeyFunc return err  ", func() {
			obj := &v1alpha1.Job{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test",
				GenerateName: "", Namespace: "namespace", SelfLink: "", UID: types.UID("xxxx"),
				ResourceVersion: "", Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ManagedFields: nil}, Spec: v1alpha1.JobSpec{},
				Status: v1alpha1.JobStatus{}}
			rs, _ := model.Factory(obj, agent.EventAdd, ctr.cacheIndexers)
			patches := gomonkey.ApplyMethod(reflect.TypeOf(new(model.VCJobModel)), "GetModelKey",
				func(_ *model.VCJobModel) string {
					return ""
				})
			defer patches.Reset()
			err := ctr.SyncHandler(rs)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey("err != nil when index getByKey return err  ", func() {
			obj := &v1alpha1.Job{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test",
				GenerateName: "", Namespace: "namespace", SelfLink: "", UID: types.UID("xxxx"),
				ResourceVersion: "", Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ManagedFields: nil}, Spec: v1alpha1.JobSpec{},
				Status: v1alpha1.JobStatus{}}
			rs, _ := model.Factory(obj, agent.EventAdd, ctr.cacheIndexers)
			rs.GetCacheIndex().Add(obj)
			patches := gomonkey.ApplyMethod(reflect.TypeOf(rs), "EventAdd", func(_ *model.VCJobModel,
				_ *agent.BusinessAgent) error {
				return nil
			})
			defer patches.Reset()
			err := ctr.SyncHandler(rs)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func newFakeController() *EventController {
	config := newTestConfig()
	kube := fake.NewSimpleClientset()
	volcano := vofake.NewSimpleClientset()
	jobInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(volcano,
		time.Second*common.InformerInterval, externalversions.WithTweakListOptions(func(options *v1.ListOptions) {
			return
		}))
	deploymentFactory := informers.NewSharedInformerFactoryWithOptions(kube, time.Second*common.InformerInterval,
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			return
		}))
	jobInformer := jobInformerFactory.Batch().V1alpha1().Jobs()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.VCJobType] = jobInformer.Informer().GetIndexer()
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()
	c, err := NewEventController(kube, volcano, config, InformerInfo{JobInformer: jobInformer,
		DeployInformer: deploymentInformer, CacheIndexers: cacheIndexer}, make(chan struct{}))
	if err != nil {
		return nil
	}
	return c
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
