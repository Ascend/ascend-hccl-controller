/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package model

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/workqueue"
	v1alpha1apis "volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	_ "hccl-controller/pkg/testtool"
)

const (
	NameSpace    = "namespace"
	Name         = "test1"
	DataKey      = "hccl.json"
	DataValue    = `{"status":"initializing"}`
	CMName       = "rings-config-test1"
	Initializing = "initializing"
)

// TestFactory test Factory
func TestFactory(t *testing.T) {
	Convey("model Factory", t, func() {
		Convey("err != nil when obj == nil", func() {
			_, err := Factory(nil, "", nil)
			So(err, ShouldNotEqual,
				nil)
		})

		Convey("err !=nil&  when obj is daemonSet ", func() {
			obj := &appsV1.DaemonSet{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test1",
				GenerateName: "", Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "",
				Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ClusterName: "", ManagedFields: nil}, Spec: appsV1.DaemonSetSpec{},
				Status: appsV1.DaemonSetStatus{}}
			_, err := Factory(obj, "add", nil)
			So(err, ShouldNotEqual, nil)
		})

		Convey("err ==nil& resourceHandle = jobHandle when obj is job ", func() {
			obj := &v1alpha1apis.Job{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test1",
				GenerateName: "", Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "",
				Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ClusterName: "", ManagedFields: nil}, Spec: v1alpha1apis.JobSpec{},
				Status: v1alpha1apis.JobStatus{}}
			rs, _ := Factory(obj, "add", nil)
			So(rs, ShouldEqual, nil)
		})

		Convey("err ==nil& resourceHandle = DeploymentHandle when obj is deployment ", func() {
			replicas := int32(1)
			obj := &appsV1.Deployment{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{Name: "test1",
				GenerateName: "", Namespace: "tt1", SelfLink: "", UID: types.UID("xxxx"), ResourceVersion: "",
				Generation: 0, CreationTimestamp: metav1.Now(), DeletionTimestamp: nil,
				DeletionGracePeriodSeconds: nil, Labels: nil, Annotations: nil, OwnerReferences: nil,
				Finalizers: nil, ClusterName: "", ManagedFields: nil},
				Spec: appsV1.DeploymentSpec{Replicas: &replicas}, Status: appsV1.DeploymentStatus{}}
			rs, _ := Factory(obj, "add", nil)
			So(rs, ShouldEqual, nil)
		})
	})
}

// TestRanktableFactory test RanktableFactory
func TestRanktableFactory(t *testing.T) {
	Convey("model RankTableFactory", t, func() {
		model := &VCJobModel{}
		Convey("err != nil when obj == nil", func() {
			patch := ApplyMethod(reflect.TypeOf(model), "GenerateGrouplist", func(_ *VCJobModel) (
				[]*v1.Group, int32, error) {
				return nil, int32(0), fmt.Errorf("test")
			})
			defer patch.Reset()
			_, _, err := RanktableFactory(model, v1.RankTableStatus{Status: ""}, "")
			So(err, ShouldNotEqual, nil)
		})

		Convey("err ==nil& when RankTableStatus is ok and version is v1", func() {
			model = &VCJobModel{taskSpec: append([]v1alpha1apis.TaskSpec(nil), v1alpha1apis.TaskSpec{})}
			patch := ApplyMethod(reflect.TypeOf(model), "GenerateGrouplist", func(_ *VCJobModel) (
				[]*v1.Group, int32, error) {
				return nil, int32(1), nil
			})
			defer patch.Reset()
			rt, _, err := RanktableFactory(model, v1.RankTableStatus{Status: Initializing}, "v1")
			So(err, ShouldEqual, nil)
			So(rt.GetStatus(), ShouldEqual, "initializing")
			rv := reflect.ValueOf(rt).Elem()
			So(rv.FieldByName("GroupCount").String(), ShouldEqual, "1")
		})

		Convey("err ==nil& when RankTableStatus is ok and version is v2", func() {
			model = &VCJobModel{taskSpec: append([]v1alpha1apis.TaskSpec(nil), v1alpha1apis.TaskSpec{})}
			pathch := ApplyMethod(reflect.TypeOf(model), "GenerateGrouplist", func(_ *VCJobModel) (
				[]*v1.Group, int32, error) {
				return nil, int32(1), nil
			})
			defer pathch.Reset()
			rt, _, err := RanktableFactory(model, v1.RankTableStatus{Status: Initializing}, "v2")
			So(err, ShouldEqual, nil)
			So(rt.GetStatus(), ShouldEqual, "initializing")
			rv := reflect.ValueOf(rt).Elem()
			So(rv.FieldByName("ServerCount").String(), ShouldEqual, "0")
		})
	})
}

// TestCheckCMCreation test CheckCMCreation
func TestCheckCMCreation(t *testing.T) {
	const (
		CmInterval = 2
		CmTimeout  = 5
	)
	config := &agent.Config{
		DryRun:           false,
		DisplayStatistic: true,
		PodParallelism:   1,
		CmCheckInterval:  CmInterval,
		CmCheckTimeout:   CmTimeout,
	}
	Convey("model checkCMCreation", t, func() {
		fakeClient := fake.NewSimpleClientset()
		fakeCoreV1 := fakeClient.CoreV1()
		cms := fakeCoreV1.ConfigMaps(NameSpace)
		Convey("err == nil when Normal", func() {
			checkCmWhenNormal(cms, fakeClient, config)
		})
		Convey("err != nil when Label not exist", func() {
			data := make(map[string]string, 1)
			label := make(map[string]string, 1)
			data[DataKey] = DataValue
			putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
				Namespace: "namespace", Labels: label}, Data: data}
			cms.Create(context.TODO(), putCM, metav1.CreateOptions{})
			getCM, err := checkCMCreation(NameSpace, Name, fakeClient, config)
			So(err, ShouldNotEqual, nil)
			So(getCM, ShouldEqual, nil)
		})
		Convey("err != nil when cm not exist", func() {
			data := make(map[string]string, 1)
			label := make(map[string]string, 1)
			data[DataKey] = DataValue
			putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "rings-config-test12",
				Namespace: "namespace", Labels: label}, Data: data}
			cms.Create(context.TODO(), putCM, metav1.CreateOptions{})
			getCM, err := checkCMCreation(NameSpace, Name, fakeClient, config)
			So(err, ShouldNotEqual, nil)
			So(getCM, ShouldEqual, nil)
		})
	})
}

func checkCmWhenNormal(cms typedcorev1.ConfigMapInterface, fakeClient *fake.Clientset, config *agent.Config) {
	data := make(map[string]string, 1)
	label := make(map[string]string, 1)
	data[DataKey] = DataValue
	label[agent.Key910] = agent.Val910
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: "namespace", Labels: label}, Data: data}
	cms.Create(context.TODO(), putCM, metav1.CreateOptions{})
	getCM, err := checkCMCreation(NameSpace, Name, fakeClient, config)
	So(err, ShouldEqual, nil)
	So(getCM.String(), ShouldEqual, putCM.String())
}

// TestVCJobModelEventAdd test VCJobModel_EventAdd
func TestVCJobModelEventAdd(t *testing.T) {
	Convey("model VCJobModel_EventAdd", t, func() {
		model := &VCJobModel{JobInfo: agent.JobInfo{JobNamespace: "namespace", JobName: "test"}}
		const (
			CmInterval = 2
			CmTimeout  = 5
			TimeSleep  = 3
		)

		config := &agent.Config{
			DryRun:           false,
			DisplayStatistic: false,
			PodParallelism:   1,
			CmCheckInterval:  CmInterval,
			CmCheckTimeout:   CmTimeout,
		}
		ag := &agent.BusinessAgent{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
				CmTimeout*time.Millisecond, TimeSleep*time.Second), "Pods"),
			KubeClientSet:  fake.NewSimpleClientset(),
			BusinessWorker: make(map[string]agent.Worker, 1),
			Config:         config,
		}
		Convey("err == nil when BusinessWorker [namespace/name] exist", func() {
			ag.BusinessWorker["namespace/test"] = nil
			err := model.EventAdd(ag)
			So(err, ShouldEqual, nil)
			So(len(ag.BusinessWorker), ShouldEqual, 1)
		})
		Convey("err !=nil&  when configmap is not exist ", func() {
			patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface, _ *agent.Config) (
				*corev1.ConfigMap, error) {
				return nil, fmt.Errorf(" failed to get configmap for job")
			})
			defer patches.Reset()
			err := model.EventAdd(ag)
			So(err, ShouldNotEqual, nil)
			So(len(ag.BusinessWorker), ShouldEqual, 0)
		})
		Convey("err !=nil & when rankTableFactory return nil", func() {
			eventAddWhenFactNil(model, ag)
		})

		Convey("err ==nil& when jobStartString is ok and version is v2", func() {
			eventAddWhenVersionV2(model, ag)
		})
	})
}

func eventAddWhenVersionV2(model *VCJobModel, ag *agent.BusinessAgent) {
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface, _ *agent.Config) (
		*corev1.ConfigMap, error) {
		data := make(map[string]string, 1)
		data[DataKey] = DataValue
		putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
			Namespace: "namespace"}, Data: data}
		return putCM, nil
	})
	defer patches.Reset()
	model = &VCJobModel{taskSpec: append([]v1alpha1apis.TaskSpec(nil), v1alpha1apis.TaskSpec{})}
	patch := ApplyMethod(reflect.TypeOf(model), "GenerateGrouplist", func(_ *VCJobModel) (
		[]*v1.Group, int32, error) {
		return nil, int32(1), nil
	})
	defer patch.Reset()
	err := model.EventAdd(ag)
	So(err, ShouldEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 1)
}

func eventAddWhenFactNil(model *VCJobModel, ag *agent.BusinessAgent) {
	patches := ApplyFunc(checkCMCreation, func(_, _ string, _ kubernetes.Interface, _ *agent.Config) (
		*corev1.ConfigMap, error) {
		data := make(map[string]string, 1)
		data[DataKey] = DataValue
		putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
			Namespace: "namespace"}, Data: data}
		return putCM, nil
	})
	defer patches.Reset()
	patches2 := ApplyFunc(RanktableFactory, func(_ ResourceEventHandler, _ v1.RankTableStatus, _ string) (
		v1.RankTabler, int32, error) {
		return nil, int32(0), fmt.Errorf("generate group list from job error")
	})
	defer patches2.Reset()
	err := model.EventAdd(ag)
	So(err, ShouldNotEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 0)
}

// TestVCJobModelEventUpdate test VCJobModel_EventUpdate
func TestVCJobModelEventUpdate(t *testing.T) {
	Convey("model VCJobModel_EventUpdate", t, func() {
		const (
			CmTimeout     = 5
			TimeSleep     = 3
			WorkLenExpect = 2
		)
		model := &VCJobModel{JobInfo: agent.JobInfo{JobNamespace: "namespace", JobName: "test"}}
		ag := &agent.BusinessAgent{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
				CmTimeout*time.Millisecond, TimeSleep*time.Second), "Pods"),
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
			patch := ApplyMethod(reflect.TypeOf(model), "EventAdd", func(vc *VCJobModel,
				agent *agent.BusinessAgent) error {
				agent.BusinessWorker[fmt.Sprintf("%s/%s", vc.JobNamespace, vc.JobName)] = nil
				return nil
			})
			defer patch.Reset()
			err := model.EventUpdate(ag)
			So(err, ShouldEqual, nil)
			So(len(ag.BusinessWorker), ShouldEqual, WorkLenExpect)
		})
		Convey("err != nil  when eventAdd has error", func() {
			updateWhenAddErr(model, ag)
		})
	})
}

func updateWhenAddErr(model *VCJobModel, ag *agent.BusinessAgent) {
	patch := ApplyMethod(reflect.TypeOf(model), "EventAdd", func(_ *VCJobModel,
		agent *agent.BusinessAgent) error {
		return fmt.Errorf("get configmap error")
	})
	defer patch.Reset()
	err := model.EventUpdate(ag)
	So(err, ShouldNotEqual, nil)
	So(len(ag.BusinessWorker), ShouldEqual, 0)
}

// TestVCJobModelGenerateGrouplist test VCJobModel_GenerateGrouplist
func TestVCJobModelGenerateGrouplist(t *testing.T) {
	Convey("model VCJobModel_GenerateGrouplist", t, func() {
		const (
			TaskRep   = 2
			RepExpect = 2
		)

		model := &VCJobModel{JobInfo: agent.JobInfo{JobNamespace: "namespace", JobName: "test"}}
		Convey("err == nil & Group is ok ", func() {
			resouceList := make(corev1.ResourceList)
			resouceList[agent.ResourceName] = *resource.NewScaledQuantity(TaskRep, 0)
			containers := []corev1.Container{
				{Resources: corev1.ResourceRequirements{Limits: resouceList}},
				{Resources: corev1.ResourceRequirements{Limits: resouceList}},
			}
			model.taskSpec = append(model.taskSpec, v1alpha1apis.TaskSpec{Replicas: TaskRep,
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: containers}}})
			groupList, re, _ := model.GenerateGrouplist()
			So(len(groupList), ShouldEqual, 1)
			So(groupList[0].DeviceCount, ShouldEqual, "8")
			So(re, ShouldEqual, RepExpect)
		})
	})
}
