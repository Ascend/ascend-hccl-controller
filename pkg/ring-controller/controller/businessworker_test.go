/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
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

// Package controller for run the logic
package controller

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"hccl-controller/pkg/ring-controller/controller/mock_kubernetes"
	"hccl-controller/pkg/ring-controller/controller/mock_v1"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_businessWorker_statistic(t *testing.T) {

	tests := []struct {
		name   string
		worker *businessWorker
	}{
		{
			name:   "test1:cachePod equals task and channel need to stop",
			worker: newMockBusinessWorkerforStatistic(1, 1, false),
		},
		{
			name:   "test2:cachePod equals task and channel already stopped ",
			worker: newMockBusinessWorkerforStatistic(1, 1, true),
		},
		{
			name:   "test3:cachePod number don't equals task",
			worker: newMockBusinessWorkerforStatistic(int32(two), 1, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.worker
			if b.statisticStopped {
				go func() {
					time.Sleep(1 * time.Second)
					b.statisticSwitch <- struct{}{}
				}()
			}
			b.statistic(twosecond)

		})
	}
}

func newMockBusinessWorkerforStatistic(cachedPodNum, taskReplicasTotal int32, statisticStopped bool) *businessWorker {
	return &businessWorker{
		statisticSwitch:      make(chan struct{}),
		podsIndexer:          nil,
		jobVersion:           1,
		jobUID:               "11",
		jobNamespace:         "vcjob",
		jobName:              "jobname",
		statisticStopped:     statisticStopped,
		cachedPodNum:         cachedPodNum,
		taskReplicasTotal:    taskReplicasTotal,
		jobCreationTimestamp: metav1.NewTime(time.Now()),
	}
}

type args struct {
	pod      *apiCoreV1.Pod
	podExist bool
	podInf   *podIdentifier
}

type testCaseForWorker struct {
	name    string
	args    args
	wantErr bool
}

func Test_businessWorker_SyncHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockCm := mock_v1.NewMockConfigMapInterface(ctrl)
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockCm.EXPECT().Update(gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockV1.EXPECT().ConfigMaps(gomock.Any()).Return(mockCm).Times(four)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(four)
	job := mockJob()
	pod := mockPod()
	tests := []testCaseForWorker{
		newTestCaseForWorker("test1: task ==0,return directly", false, nil, false,
			nil),
		newTestCaseForWorker("test2: dryrun case,return directly", false, nil, false,
			mockPodIdentify("delete")),
		newTestCaseForWorker("test3:undefined situation, return directly ", false, nil,
			false, mockPodIdentify("add")),
		newTestCaseForWorker("test4:add pod and no error returned", false, pod, true,
			mockPodIdentify("add")),
		newTestCaseForWorker("test5:delete pod and no error returned", false, pod, true,
			mockPodIdentify("delete")),
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBusinessWorker(mockK8s, nil, nil, false, job)
			if i == 0 {
				b.cachedPodNum = 0
				b.taskReplicasTotal = 0
			}
			if i == 1 {
				b.dryRun = true
			}
			if err := b.syncHandler(tt.args.pod, tt.args.podExist, tt.args.podInf); (err != nil) != tt.wantErr {
				t.Errorf("syncHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newTestCaseForWorker(name string, wantErr bool, pod *apiCoreV1.Pod,
	podExist bool, podIdentifier *podIdentifier) testCaseForWorker {
	return testCaseForWorker{
		name:    name,
		wantErr: wantErr,
		args: args{
			pod:      pod,
			podExist: podExist,
			podInf:   podIdentifier,
		},
	}
}

func mockPodIdentify(event string) *podIdentifier {
	return &podIdentifier{
		namespace: "vcjob",
		name:      "hccl-test",
		jobName:   "jobname",
		eventType: event,
	}
}

func generateInstance(worker *businessWorker) {
	var instance Instance
	deviceInfo := mockJSON()
	err := json.Unmarshal([]byte(deviceInfo), &instance)
	if err != nil {
		return
	}
	worker.configmapData.GroupList[0].InstanceList = append(worker.configmapData.GroupList[0].InstanceList, &instance)
}

func mockJSON() string {
	var deviceInfo = map[string]interface{}{
		"pod_name":  "hccl-test",
		"server_id": "51.38.60.7",
		"devices": []map[string]string{
			{
				"device_id": "0",
				"device_ip": "192.168.100.100",
			},
		},
	}
	bytes, err := json.Marshal(deviceInfo)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func Test_businessWorker_handleDeleteEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockCm := mock_v1.NewMockConfigMapInterface(ctrl)
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockCm.EXPECT().Update(gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockV1.EXPECT().ConfigMaps(gomock.Any()).Return(mockCm).Times(four)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(four)
	job := mockJob()
	tests := []struct {
		name    string
		podInf  *podIdentifier
		wantErr bool
		worker  *businessWorker
	}{
		{
			name:    "test1:no data to remove, return directly",
			podInf:  mockPodIdentify("delete"),
			wantErr: false,
			worker:  newMockBusinessWorkerforStatistic(1, 1, true),
		},
		{
			name:    "test2:normal situation, no error",
			podInf:  mockPodIdentify("delete"),
			wantErr: false,
			worker:  newMockBusinessWorkerforStatistic(1, 1, true),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBusinessWorker(mockK8s, nil, nil, false, job)
			if i == 0 {
				generateInstance(b)
			}
			if err := b.handleDeleteEvent(tt.podInf); (err != nil) != tt.wantErr {
				t.Errorf("handleDeleteEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_businessWorker_handleAddUpdateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockCm := mock_v1.NewMockConfigMapInterface(ctrl)
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockCm.EXPECT().Update(gomock.Any()).Return(mockConfigMap(), nil).Times(two)
	mockV1.EXPECT().ConfigMaps(gomock.Any()).Return(mockCm).Times(four)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(four)
	job := mockJob()
	pod := mockPod()
	var pods []*apiCoreV1.Pod
	pods = append(pods, pod.DeepCopy())
	pod.Annotations[PodDeviceKey] = mockJSON()
	pods = append(pods, pod.DeepCopy())
	pod.Name = "test"
	pods = append(pods, pod.DeepCopy())
	pod.Annotations[PodDeviceKey] = "xxx"
	pods = append(pods, pod.DeepCopy())
	tests := []testCaseForWorker{
		newTestCaseForWorker("test1:no device info,run cachezeroPodInfo,no error", false,
			pods[0], true, mockPodIdentify("add")),
		newTestCaseForWorker("test2: already cache pod info, do nothing", false,
			pods[1], true, mockPodIdentify("add")),
		newTestCaseForWorker("test3: with device info,run cachePodInfo,no error", false,
			pods[2], true, mockPodIdentify("add")),
		newTestCaseForWorker("test4: json format error", true,
			pods[3], true, mockPodIdentify("add")),
	}
	b := newBusinessWorker(mockK8s, nil, nil, false, job)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := b.handleAddUpdateEvent(tt.args.podInf, tt.args.pod); (err != nil) != tt.wantErr {
				t.Errorf("handleAddUpdateEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
