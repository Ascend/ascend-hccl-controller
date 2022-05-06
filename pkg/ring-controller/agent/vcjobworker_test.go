/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package agent

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	fakecm "k8s.io/client-go/kubernetes/typed/core/v1/fake"

	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	ranktablev2 "hccl-controller/pkg/ring-controller/ranktable/v2"
)

const (
	NameSpace = "namespace"
	DataKey   = "hccl.json"
	DataValue = `{"status":"initializing"}`
	CMName    = "rings-config-test1"
)

// TestUpdateWithFinish test UpdateWithFinish
func TestUpdateWithFinish(t *testing.T) {
	convey.Convey("agent UpdateWithFinish", t, func() {
		worker := &WorkerInfo{}
		const (
			TaskRep = 2
		)

		convey.Convey(" err == nil when tableConstructionFinished return false", func() {
			worker.taskReplicasTotal = TaskRep
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey(" err != nil when endRankTableConstruction return err", func() {
			patches := gomonkey.ApplyFunc(updateConfigMap, func(_ *WorkerInfo, _ string) error {
				return fmt.Errorf("failed to update ConfigMap for Job")
			})
			defer patches.Reset()
			worker.configmapData = &ranktablev2.RankTable{RankTableStatus: ranktablev1.
				RankTableStatus{Status: "start"}}
			worker.taskReplicasTotal = 1
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey(" err == nil when endRankTableConstruction return nil", func() {
			patches := gomonkey.ApplyFunc(updateConfigMap, func(_ *WorkerInfo, _ string) error {
				return nil
			})
			defer patches.Reset()
			worker.configmapData = &ranktablev2.RankTable{RankTableStatus: ranktablev1.
				RankTableStatus{Status: "start"}}
			worker.taskReplicasTotal = 1
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

// TestGetWorkName test GetWorkName
func TestGetWorkName(t *testing.T) {
	convey.Convey("agent GetWorkName", t, func() {
		labels := make(map[string]string, 1)

		convey.Convey(" return volcano-job when label contains VolcanoJobNameKey ", func() {
			labels[VolcanoJobNameKey] = VolcanoJobNameKey
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, VolcanoJobNameKey)
		})
		convey.Convey("  return deployment-name when label contains VolcanoJobNameKey ", func() {
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, DeploymentNameKey)
		})
	})
}

// TestUpdateConfigMap test UpdateConfigMap
func TestUpdateConfigMap(t *testing.T) {
	convey.Convey("agent updateConfigMap", t, func() {
		kube := fake.NewSimpleClientset()
		work := &WorkerInfo{kubeclientset: kube, configmapName: CMName}
		convey.Convey(" return err != nil when  cm not exist ", func() {
			err := updateConfigMap(work, NameSpace)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey(" return err != nil when label in  cm not exist Key910 ", func() {
			data := make(map[string]string, 1)
			putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
				Namespace: NameSpace}, Data: data}
			kube.CoreV1().ConfigMaps(NameSpace).Create(context.TODO(), putCM, metav1.CreateOptions{})
			err := updateConfigMap(work, NameSpace)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey(" return err != nil when update cm error ", func() {
			updateWhenUpdateCmErr(kube, work)
		})
		convey.Convey(" return err == nil when label in  cm normal ", func() {
			updateWhenCMNormal(kube, work)
		})
	})

}

func updateWhenCMNormal(kube *fake.Clientset, work *WorkerInfo) {
	data := make(map[string]string, 1)
	label := make(map[string]string, 1)
	data[DataKey] = DataValue
	label[Key910] = Val910
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: NameSpace, Labels: label}, Data: data}
	kube.CoreV1().ConfigMaps(NameSpace).Create(context.TODO(), putCM, metav1.CreateOptions{})
	work.configmapData = &ranktablev1.RankTable{RankTableStatus: ranktablev1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	err := updateConfigMap(work, NameSpace)
	convey.So(err, convey.ShouldEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(context.TODO(), CMName,
		metav1.GetOptions{})
	convey.So(cm.Data[DataKey], convey.ShouldEqual, `{"status":"completed","group_list":null,"group_count":""}`)
}

func updateWhenUpdateCmErr(kube *fake.Clientset, work *WorkerInfo) {
	label := make(map[string]string, 1)
	label[Key910] = Val910
	data := make(map[string]string, 1)
	data[DataKey] = DataValue
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: NameSpace, Labels: label}, Data: data}
	kube.CoreV1().ConfigMaps("namespace").Create(context.TODO(), putCM, metav1.CreateOptions{})
	work.configmapData = &ranktablev1.RankTable{RankTableStatus: ranktablev1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(kube.CoreV1().ConfigMaps(NameSpace)),
		"Update", func(_ *fakecm.FakeConfigMaps, _ context.Context, _ *corev1.ConfigMap,
			_ metav1.UpdateOptions) (*corev1.ConfigMap, error) {
			return nil, fmt.Errorf("update config error")
		})
	defer patch.Reset()
	err := updateConfigMap(work, NameSpace)
	convey.So(err, convey.ShouldNotEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(context.TODO(), CMName,
		metav1.GetOptions{})
	convey.So(cm.Data[DataKey], convey.ShouldEqual, DataValue)
}

// TestWorkerInfoCloseStatistic test WorkerInfo_CloseStatistic
func TestWorkerInfoCloseStatistic(t *testing.T) {
	convey.Convey("agent TestWorkerInfo_CloseStatistic", t, func() {
		w := &WorkerInfo{statisticStopped: true, statisticSwitch: make(chan struct{})}

		convey.Convey(" chan not close when statisticStopped is true ", func() {
			w.CloseStatistic()
			go func() {
				w.statisticSwitch <- struct{}{}
			}()
			_, open := <-w.statisticSwitch
			convey.So(open, convey.ShouldEqual, true)
		})

	})
}

// TestVCJobWorkerStatistic test VCJobWorker_Statistic
func TestVCJobWorkerStatistic(t *testing.T) {
	convey.Convey("agent VCJobWorker_Statistic", t, func() {
		vc := &VCJobWorker{WorkerInfo: WorkerInfo{statisticSwitch: make(chan struct{}), statisticStopped: false}}
		const (
			TaskRep   = 2
			SleepTime = 3
		)

		convey.Convey(" chan will return when chan close ", func() {
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = 1
			go func() {
				time.Sleep(SleepTime * time.Second)
				vc.CloseStatistic()
			}()
			vc.Statistic(1 * time.Second)
		})

		convey.Convey(" chan will return when taskReplicasTotal==cachedPodNum ", func() {
			const CachePod = 2
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = CachePod
			vc.Statistic(1 * time.Second)
		})
	})
}
