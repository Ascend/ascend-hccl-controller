/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */
package model

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"

	"hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// TestDeployModelEventAdd test Dep6loyModel_EventAdd
func TestDeployModelEventAdd(t *testing.T) {
	Convey("model DeployModel_EventAdd", t, func() {
		model := &DeployModel{DeployInfo: agent.DeployInfo{DeployNamespace: "namespace", DeployName: "test"}}
		const (
			CmIntervals = 2
			CmTimeout   = 5
			SleepTime   = 3
		)
		config := &agent.Config{
			CmCheckInterval:  CmIntervals,
			CmCheckTimeout:   CmTimeout,
			DryRun:           false,
			DisplayStatistic: false,
			PodParallelism:   1,
		}
		ag := &agent.BusinessAgent{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
				CmTimeout*time.Millisecond, SleepTime*time.Second), "Pods"),
			BusinessWorker: make(map[string]agent.Worker, 1),
			Config:         config,
		}

		Convey("err !=nil&  when configmap is not exist ", func() {
			eventAddWhenCMNotExist(model, ag)
		})
		Convey("err !=nil & when rankTableFactory return nil", func() {
			eventAddWhenFacNil(model, ag)
		})

		Convey("err ==nil& when jobStartString is ok and version is v2", func() {
			eventAddWhenV2(model, ag)
		})

		Convey("err == nil when BusinessWorker [namespace/name] exist", func() {
			eventAddWhenWorkerExist(ag, model)
		})
	})
}

func eventAddWhenWorkerExist(ag *agent.BusinessAgent, model *DeployModel) {
	ag.BusinessWorker["namespace/test"] = nil
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface,
		_ *agent.Config) (*corev1.ConfigMap, error) {
		data := make(map[string]string, 1)
		data[DataKey] = DataValue
		putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
			Namespace: "namespace"}, Data: data}
		return putCM, nil
	})
	defer patches.Reset()
	patch := ApplyFunc(RanktableFactory, func(_ ResourceEventHandler, _ v1.RankTableStatus, _ string) (
		v1.RankTabler, int32, error) {
		return nil, int32(1), nil
	})
	defer patch.Reset()
	err := model.EventAdd(ag)
	So(err, ShouldEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 1)
}

func eventAddWhenV2(model *DeployModel, ag *agent.BusinessAgent) {
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface,
		_ *agent.Config) (*corev1.ConfigMap, error) {
		data := make(map[string]string, 1)
		data[DataKey] = DataValue
		putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
			Namespace: NameSpace}, Data: data}
		return putCM, nil
	})
	defer patches.Reset()
	model = &DeployModel{}
	patch := ApplyFunc(RanktableFactory, func(_ ResourceEventHandler, _ v1.RankTableStatus, _ string) (v1.RankTabler,
		int32, error) {
		return nil, int32(1), nil
	})
	defer patch.Reset()
	err := model.EventAdd(ag)
	So(err, ShouldEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 1)
}

func eventAddWhenFacNil(model *DeployModel, ag *agent.BusinessAgent) {
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface,
		_ *agent.Config) (*corev1.ConfigMap, error) {
		data := make(map[string]string, 1)
		data[DataKey] = DataValue
		putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
			Namespace: NameSpace}, Data: data}
		return putCM, nil
	})
	defer patches.Reset()
	patches2 := ApplyFunc(RanktableFactory, func(_ ResourceEventHandler,
		_ v1.RankTableStatus, _ string) (v1.RankTabler, int32, error) {
		return nil, int32(0), errors.New("generated group list from job error")
	})
	defer patches2.Reset()
	err := model.EventAdd(ag)
	So(err, ShouldNotEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 0)
}

func eventAddWhenCMNotExist(model *DeployModel, ag *agent.BusinessAgent) {
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface,
		_ *agent.Config) (*corev1.ConfigMap, error) {

		return nil, fmt.Errorf(" failed to get configmap for job")
	})
	defer patches.Reset()
	err := model.EventAdd(ag)
	So(len(ag.BusinessWorker), ShouldEqual, 0)
	So(err, ShouldNotEqual, nil)
}

// TestDeployModelEventUpdate test DeployModel_EventUpdate
func TestDeployModelEventUpdate(t *testing.T) {
	const (
		WorkLenExpect = 2
		CmTimeout     = 5
		SleepTime     = 3
	)
	Convey("model DeployModel_EventUpdate", t, func() {
		model := &DeployModel{DeployInfo: agent.DeployInfo{DeployNamespace: "namespace", DeployName: "test"}}
		ag := &agent.BusinessAgent{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
				CmTimeout*time.Millisecond, SleepTime*time.Second), "Pods"),
			BusinessWorker: make(map[string]agent.Worker, 1),
		}
		Convey("err == nil when BusinessWorker exist job", func() {
			ag.BusinessWorker["namespace/test"] = nil
			err := model.EventUpdate(ag)
			So(err, ShouldEqual, nil)
			So(len(ag.BusinessWorker), ShouldEqual, 1)
		})

		Convey("err == nil && len(map)==len(map)+1 when BusinessWorker do not exist job", func() {
			ag.BusinessWorker["namespace/test1"] = nil
			patch := ApplyMethod(reflect.TypeOf(model), "EventAdd", func(dp *DeployModel,
				agent *agent.BusinessAgent) error {
				agent.BusinessWorker[fmt.Sprintf("%s/%s", dp.DeployNamespace, dp.DeployName)] = nil
				return nil
			})
			defer patch.Reset()
			err := model.EventUpdate(ag)
			So(err, ShouldEqual, nil)
			So(len(ag.BusinessWorker), ShouldEqual, WorkLenExpect)
		})
		Convey("err != nil  when eventAdd has error", func() {
			patch := ApplyMethod(reflect.TypeOf(model), "EventAdd", func(_ *DeployModel,
				agent *agent.BusinessAgent) error {
				return fmt.Errorf("get configmap errors")
			})
			err := model.EventUpdate(ag)
			defer patch.Reset()
			So(len(ag.BusinessWorker), ShouldEqual, 0)
			So(err, ShouldNotEqual, nil)
		})
	})
}

// TestDeployModelGenerateGrouplist test DeployModel_GenerateGrouplis
func TestDeployModelGenerateGrouplist(t *testing.T) {
	Convey("model DeployModel_GenerateGrouplist", t, func() {
		const (
			WorkLenExpect = 2
			DeployRep     = 2
		)
		model := &DeployModel{replicas: DeployRep}
		Convey("err == nil & Group is ok ", func() {
			resouceList := make(corev1.ResourceList, 1)
			resouceList[agent.ResourceName] = *resource.NewScaledQuantity(2, 0)
			containers := []corev1.Container{
				{Resources: corev1.ResourceRequirements{Limits: resouceList}},
				{Resources: corev1.ResourceRequirements{Limits: resouceList}},
			}
			model.containers = containers
			groupList, re, _ := model.GenerateGrouplist()
			So(len(groupList), ShouldEqual, 1)
			So(groupList[0].DeviceCount, ShouldEqual, "8")
			So(re, ShouldEqual, WorkLenExpect)
		})
	})
}
