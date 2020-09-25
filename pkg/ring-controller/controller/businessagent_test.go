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

// Package controller for run the logic
package controller

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"hccl-controller/pkg/ring-controller/controller/mock_cache"
	"hccl-controller/pkg/ring-controller/controller/mock_kubernetes"
	"hccl-controller/pkg/ring-controller/controller/mock_v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"testing"
	"time"
	"volcano.sh/volcano/pkg/apis/batch/v1alpha1"
)

func Test_businessAgent_deleteBusinessWorker(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		namespace string
		podName   string
		worker    *businessAgent
	}{
		{
			name:      "test1:worker not exist",
			wantErr:   false,
			namespace: "vcjob",
			podName:   "hccl-test",
			worker:    createAgent(true),
		},
		{
			name:      "test1:worker exist",
			wantErr:   false,
			namespace: "vcjob",
			podName:   "hccl-test",
			worker:    createAgent(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.worker.dryRun {
				tt.worker.businessWorker["vcjob/hccl-test"] = newMockBusinessWorkerforStatistic(1, 1, false)
			}
			if err := tt.worker.DeleteBusinessWorker(tt.namespace, tt.podName); (err != nil) != tt.wantErr {
				t.Errorf("deleteBusinessWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_businessAgent_isBusinessWorkerExist(t *testing.T) {
	tests := []struct {
		name      string
		expect    bool
		namespace string
		podName   string
		worker    *businessAgent
	}{
		{
			name:      "test1:worker not exist,return directly",
			expect:    false,
			namespace: "vcjob",
			podName:   "hccl-test",
			worker:    createAgent(true),
		},
		{
			name:      "test1:worker exist,delete ok",
			expect:    true,
			namespace: "vcjob",
			podName:   "hccl-test",
			worker:    createAgent(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.worker.dryRun {
				tt.worker.businessWorker["vcjob/hccl-test"] = newMockBusinessWorkerforStatistic(1, 1, false)
			}
			tt.worker.IsBusinessWorkerExist(tt.namespace, tt.podName)
			assert.Equal(t, !tt.worker.dryRun, tt.expect)
		})
	}
}

func createAgent(dryrun bool) *businessAgent {
	return &businessAgent{
		informerFactory: nil,
		podInformer:     nil,
		podsIndexer:     nil,
		kubeClientSet:   nil,
		businessWorker:  make(map[string]*businessWorker),
		agentSwitch:     nil,
		recorder:        nil,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
			retryMilliSecond*time.Millisecond, threeMinutes*time.Second), "Pods"),
		dryRun:           dryrun,
		displayStatistic: true,
		cmCheckInterval:  decimal,
		cmCheckTimeout:   decimal,
	}
}

func createAgentForController(dryrun bool) *businessAgent {
	return &businessAgent{
		informerFactory:  nil,
		podInformer:      nil,
		podsIndexer:      nil,
		kubeClientSet:    nil,
		businessWorker:   make(map[string]*businessWorker),
		agentSwitch:      nil,
		recorder:         nil,
		workqueue:        nil,
		dryRun:           dryrun,
		displayStatistic: true,
		cmCheckInterval:  decimal,
		cmCheckTimeout:   decimal,
	}
}

func Test_businessAgent_createBusinessWorker(t *testing.T) {
	tests := []struct {
		name   string
		worker *businessAgent
	}{
		{
			name:   "test1:worker not exist,return directly",
			worker: createAgent(true),
		},
		{
			name:   "test1:worker exist,create ok",
			worker: createAgent(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.worker.dryRun {
				tt.worker.businessWorker["vcjob/hccl-test"] = newMockBusinessWorkerforStatistic(1, 1, false)
			}
			tt.worker.CreateBusinessWorker(mockJob())
		})
	}
}

func mockJob() *v1alpha1.Job {
	return &v1alpha1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch.volcano.sh/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "resnet",
			Namespace: "vc-job",
		},
		Spec: v1alpha1.JobSpec{
			SchedulerName:     "volcano",
			MinAvailable:      1,
			Queue:             "default",
			MaxRetry:          three,
			PriorityClassName: "",
			Tasks:             mockTask(),
		},
		Status: v1alpha1.JobStatus{},
	}
}

func mockTask() []v1alpha1.TaskSpec {
	return []v1alpha1.TaskSpec{
		{
			Name:     "default-test",
			Replicas: 1,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "",
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									ResourceName: resource.MustParse("1"),
								},
								Requests: v1.ResourceList{
									ResourceName: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
		}}
}

func mockPod() *v1.Pod {
	localTime, err := time.Parse("2006-01-02 15:04:05", "2017-04-11 13:33:37")
	if err != nil {
		return nil
	}
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch.volcano.sh/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "resnet",
			Namespace:         "vc-job",
			CreationTimestamp: metav1.NewTime(localTime),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "",
					Kind:               "",
					Name:               "",
					UID:                "10",
					Controller:         nil,
					BlockOwnerDeletion: nil,
				},
			},
			Annotations: map[string]string{
				PodGroupKey: "default-test",
			},
		},
		Spec:   mockSpec(),
		Status: v1.PodStatus{},
	}
}

func mockSpec() v1.PodSpec {
	return v1.PodSpec{
		Containers: []v1.Container{
			{
				Image: "",
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceName: resource.MustParse("1"),
					},
					Requests: v1.ResourceList{
						ResourceName: resource.MustParse("1"),
					},
				},
			},
		},
	}
}
func Test_businessAgent_doWork(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockIndexer := mock_cache.NewMockIndexer(ctrl)
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(nil, false, fmt.Errorf("mock error"))
	pod := mockPod()
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	pod.OwnerReferences[0].Name = "jobname"
	pod.OwnerReferences[0].UID = "11"
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	pod.Annotations[PodJobVersion] = "0"
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	pod.Annotations[PodJobVersion] = "2"
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	pod.Annotations[PodJobVersion] = "1"
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	pod.Annotations[PodDeviceKey] = "{\"pod_name\":\"0\",\"server_id\":\"51.38.60.7\",\"devices\":[{\"device_id\":\"0\",\"device_ip\":\"192.168.100.100\"}]}\n"
	mockIndexer.EXPECT().GetByKey(gomock.Any()).Return(pod.DeepCopy(), true, nil)
	tests := []testCase{
		getTestCaseForDoWork("test0：precheck failed", false, true, false),
		getTestCaseForDoWork("test1：should split key error", "vcjob/hccl-test", true, false),
		getTestCaseForDoWork("test2：no pod from listener", "vcjob/hccl-test/jobname/add", true, false),
		getTestCaseForDoWork("test3：worker not exist", "vcjob/hccl-test/jobname/add", false, false),
		getTestCaseForDoWork("test4：worker exist but OwnerReferences check fail", "vcjob/hccl-test/jobname/add", true, true),
		getTestCaseForDoWork("test5：worker exist, version check error", "vcjob/hccl-test/jobname/add", true, true),
		getTestCaseForDoWork("test6：worker exist.pod version < worker version", "vcjob/hccl-test/jobname/add", true, true),
		getTestCaseForDoWork("test7：worker exist,Pod version > worker version", "vcjob/hccl-test/jobname/add", false, true),
		getTestCaseForDoWork("test8：worker exist,Pod version = worker version", "vcjob/hccl-test/jobname/add", false, true),
		getTestCaseForDoWork("test9：worker exist,Pod have device info", "vcjob/hccl-test/jobname/add", true, true),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.workAgent.podsIndexer = mockIndexer
			if tt.worker {
				tt.workAgent.businessWorker["vcjob/jobname"] = newMockBusinessWorkerforStatistic(1, 1, false)
			}
			if got := tt.workAgent.doWork(tt.obj); got != tt.want {
				t.Errorf("doWork() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testCase struct {
	obj       interface{}
	name      string
	workAgent *businessAgent
	want      bool
	worker    bool
}

func getTestCaseForDoWork(name string, obj interface{}, want, worker bool) testCase {
	return testCase{
		name:      name,
		obj:       obj,
		workAgent: createAgent(false),
		want:      want,
		worker:    worker,
	}
}

func Test_businessAgent_enqueuePod(t *testing.T) {
	tests := []struct {
		name      string
		obj       interface{}
		eventType string
	}{
		{
			name:      "test1: Pod be added to queue",
			obj:       mockPod(),
			eventType: "add",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := createAgent(false)
			b.enqueuePod(tt.obj, tt.eventType)
			obj, _ := b.workqueue.Get()
			assert.NotNil(t, obj)

		})
	}
}

func Test_businessAgent_CheckConfigmapCreation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockCm := mock_v1.NewMockConfigMapInterface(ctrl)
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockConfigMap(), nil)
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock error"))
	cm := mockConfigMap()
	cm.ObjectMeta.Labels[Key910] = "ascend=310"
	mockCm.EXPECT().Get(gomock.Any(), gomock.Any()).Return(cm, nil)
	mockV1.EXPECT().ConfigMaps(gomock.Any()).Return(mockCm).Times(three)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(three)
	tests := []struct {
		name    string
		job     *v1alpha1.Job
		want    *v1.ConfigMap
		wantErr bool
	}{
		{
			name:    "test",
			job:     mockJob(),
			want:    mockConfigMap(),
			wantErr: false,
		},
		{
			name:    "test2: get configmap error and return error",
			job:     mockJob(),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "test3: configmap invalid and return error",
			job:     mockJob(),
			want:    nil,
			wantErr: true,
		},
	}
	b := createAgent(false)
	b.kubeClientSet = mockK8s
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := b.CheckConfigmapCreation(tt.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckConfigmapCreation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
