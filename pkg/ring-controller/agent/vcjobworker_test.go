/*
 * Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

package agent

import (
	"fmt"
	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	v2 "hccl-controller/pkg/ring-controller/ranktable/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	fakecm "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"reflect"
	"testing"
	"time"
)

const (
	NameSpace = "namespace"
	Name      = "test1"
	DataKey   = "hccl.json"
	DataValue = "{\"status\":\"initializing\"}"
	CMName    = "rings-config-test1"
)

// TestUpdateWithFinish test UpdateWithFinish
func TestUpdateWithFinish(t *testing.T) {
	Convey("agent UpdateWithFinish", t, func() {
		worker := &WorkerInfo{}
		const (
			TaskRep = 2
		)

		Convey(" err == nil when tableConstructionFinished return false", func() {
			worker.taskReplicasTotal = TaskRep
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			So(err, ShouldEqual, nil)
		})
		Convey(" err != nil when endRankTableConstruction return err", func() {
			patches := ApplyFunc(updateConfigMap, func(_ *WorkerInfo, _ string) error {
				return fmt.Errorf("failed to update ConfigMap for Job")
			})
			defer patches.Reset()
			worker.configmapData = &v2.RankTable{RankTableStatus: v1.RankTableStatus{Status: "start"}}
			worker.taskReplicasTotal = 1
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			So(err, ShouldNotEqual, nil)
		})
		Convey(" err == nil when endRankTableConstruction return nil", func() {
			patches := ApplyFunc(updateConfigMap, func(_ *WorkerInfo, _ string) error {
				return nil
			})
			defer patches.Reset()
			worker.configmapData = &v2.RankTable{RankTableStatus: v1.RankTableStatus{Status: "start"}}
			worker.taskReplicasTotal = 1
			worker.cachedPodNum = 1
			err := updateWithFinish(worker, NameSpace)
			So(err, ShouldEqual, nil)
		})
	})
}

// TestGetWorkName test GetWorkName
func TestGetWorkName(t *testing.T) {
	Convey("agent GetWorkName", t, func() {
		labels := make(map[string]string, 1)

		Convey(" return volcano-job when label contains VolcanoJobNameKey ", func() {
			labels[VolcanoJobNameKey] = VolcanoJobNameKey
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			So(work, ShouldEqual, VolcanoJobNameKey)
		})
		Convey("  return deployment-name when label contains VolcanoJobNameKey ", func() {
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			So(work, ShouldEqual, DeploymentNameKey)
		})
	})
}

// TestUpdateConfigMap test UpdateConfigMap
func TestUpdateConfigMap(t *testing.T) {
	Convey("agent updateConfigMap", t, func() {
		kube := fake.NewSimpleClientset()
		work := &WorkerInfo{kubeclientset: kube, configmapName: CMName}
		Convey(" return err != nil when  cm not exist ", func() {
			err := updateConfigMap(work, NameSpace)
			So(err, ShouldNotEqual, nil)
		})
		Convey(" return err != nil when label in  cm not exist Key910 ", func() {
			data := make(map[string]string, 1)
			putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
				Namespace: NameSpace}, Data: data}
			kube.CoreV1().ConfigMaps(NameSpace).Create(putCM)
			err := updateConfigMap(work, NameSpace)
			So(err, ShouldNotEqual, nil)
		})
		Convey(" return err != nil when update cm error ", func() {
			updateWhenUpdateCmErr(kube, work)
		})
		Convey(" return err == nil when label in  cm normal ", func() {
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
	kube.CoreV1().ConfigMaps(NameSpace).Create(putCM)
	work.configmapData = &v1.RankTable{RankTableStatus: v1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	err := updateConfigMap(work, NameSpace)
	So(err, ShouldEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(CMName,
		metav1.GetOptions{})
	So(cm.Data[DataKey], ShouldEqual, "{\"status\":\"completed\","+
		"\"group_list\":null,\"group_count\":\"\"}")
}

func updateWhenUpdateCmErr(kube *fake.Clientset, work *WorkerInfo) {
	label := make(map[string]string, 1)
	label[Key910] = Val910
	data := make(map[string]string, 1)
	data[DataKey] = DataValue
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: NameSpace, Labels: label}, Data: data}
	kube.CoreV1().ConfigMaps("namespace").Create(putCM)
	work.configmapData = &v1.RankTable{RankTableStatus: v1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	patch := ApplyMethod(reflect.TypeOf(kube.CoreV1().ConfigMaps(NameSpace)),
		"Update", func(_ *fakecm.FakeConfigMaps, _ *corev1.ConfigMap) (*corev1.ConfigMap, error) {
			return nil, fmt.Errorf("update config error")
		})
	defer patch.Reset()
	err := updateConfigMap(work, NameSpace)
	So(err, ShouldNotEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(CMName,
		metav1.GetOptions{})
	So(cm.Data[DataKey], ShouldEqual, DataValue)
}

// TestWorkerInfoCloseStatistic test WorkerInfo_CloseStatistic
func TestWorkerInfoCloseStatistic(t *testing.T) {
	Convey("agent TestWorkerInfo_CloseStatistic", t, func() {
		w := &WorkerInfo{statisticStopped: true, statisticSwitch: make(chan struct{})}

		Convey(" chan not close when statisticStopped is true ", func() {
			w.CloseStatistic()
			go func() {
				w.statisticSwitch <- struct{}{}
			}()
			_, open := <-w.statisticSwitch
			So(open, ShouldEqual, true)
		})

	})
}

// TestVCJobWorkerStatistic test VCJobWorker_Statistic
func TestVCJobWorkerStatistic(t *testing.T) {
	Convey("agent VCJobWorker_Statistic", t, func() {
		vc := &VCJobWorker{WorkerInfo: WorkerInfo{statisticSwitch: make(chan struct{}), statisticStopped: false}}
		const (
			TaskRep   = 2
			SleepTime = 3
		)

		Convey(" chan will return when chan close ", func() {
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = 1
			go func() {
				time.Sleep(SleepTime * time.Second)
				vc.CloseStatistic()
			}()
			vc.Statistic(1 * time.Second)
		})

		Convey(" chan will return when taskReplicasTotal==cachedPodNum ", func() {
			const CachePod = 2
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = CachePod
			vc.Statistic(1 * time.Second)
		})
	})
}
