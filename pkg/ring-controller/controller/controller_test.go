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
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"hccl-controller/pkg/ring-controller/controller/mock_cache"
	"hccl-controller/pkg/ring-controller/controller/mock_controller"
	"hccl-controller/pkg/ring-controller/controller/mock_kubernetes"
	"hccl-controller/pkg/ring-controller/controller/mock_v1"
	"hccl-controller/pkg/ring-controller/controller/mock_v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"reflect"
	"testing"
	"time"
	"volcano.sh/volcano/pkg/apis/batch/v1alpha1"
)

// Test_startPerformanceMonitorServer test method startPerformanceMonitorServer
func Test_startPerformanceMonitorServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Http check ,response should be 200",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go startPerformanceMonitorServer()
			url := "http://localhost:6060/debug/pprof"
			response, err := http.Get(url)
			if err != nil || response == nil || response.Body == nil {
				t.Fatalf("check performance server failed,%v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != status {
				t.Errorf("check performance server failed,%v", err)
			}

		})
	}
}

// TestController_createBusinessWorker  test  method createBusinessWorker
func TestController_createBusinessWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := mock_controller.NewMockWorkAgentInterface(ctrl)
	mockAgent.EXPECT().CreateBusinessWorker(gomock.Any()).Return(nil)
	tests := []struct {
		configMap *v1.ConfigMap
		name      string
		wantErr   bool
	}{
		{
			name:    "test1:return error for json format error",
			wantErr: true,
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					ConfigmapKey: "XXX",
				},
			},
		},
		{
			name:    "test2:return error for cm status error",
			wantErr: true,
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					ConfigmapKey: "{\"status\":\"error\"}",
				},
			},
		},
		{
			name:      "test3:normal situation, no errors",
			wantErr:   false,
			configMap: mockConfigMap(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAgent.EXPECT().CheckConfigmapCreation(gomock.Any()).Return(tt.configMap, nil)
			c := &Controller{
				workAgentInterface: mockAgent,
			}
			if err := c.createBusinessWorker(&v1alpha1.Job{}); (err != nil) != tt.wantErr {
				t.Errorf("createBusinessWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestController_syncHandler  test method  syncHandler
func TestController_syncHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAgent := mock_controller.NewMockWorkAgentInterface(ctrl)
	mockAgent.EXPECT().CreateBusinessWorker(gomock.Any()).Return(nil).Times(three)
	mockAgent.EXPECT().CheckConfigmapCreation(gomock.Any()).Return(mockConfigMap(), nil).Times(three)
	mockIndexr := mock_cache.NewMockIndexer(ctrl)
	mockIndexr.EXPECT().GetByKey(gomock.Any()).Return(mockJob(), true, nil).Times(four)
	mockIndexr.EXPECT().GetByKey(gomock.Any()).Return(nil, false, fmt.Errorf("mock error"))
	mockIndexr.EXPECT().GetByKey(gomock.Any()).Return(nil, false, nil) // no pod existed

	tests := []controllerTestCase{
		mackTestCase("test1: Key format error, should return error", "vcjob/testpod/delete/error", true),
		mackTestCase("test2:add event,no err returned", "vcjob/testpod/add", false),
		mackTestCase("test3: update event,no err returned", "vcjob/testpod/update", false),
		mackTestCase("test4: delete event but indexer have pod ,should return error", "vcjob/pod1/delete", true),
		mackTestCase("test5: Unfinded situation, should return error", "vcjob/testpod/new", true),
		mackTestCase("test6: Indexer return error, should return error", "vcjob/pod2/delete", true),
		mackTestCase("test7: delete event,no err returned", "vcjob/pod3/delete", false),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				workAgentInterface: mockAgent,
				jobsIndexer:        mockIndexr,
				businessAgent:      createAgent(true),
			}
			if err := c.syncHandler(tt.key); (err != nil) != tt.wantErr {
				t.Errorf("syncHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type controllerTestCase struct {
	name    string
	key     string
	wantErr bool
}

func mackTestCase(name, key string, wantErr bool) controllerTestCase {
	return controllerTestCase{
		name:    name,
		key:     key,
		wantErr: wantErr,
	}
}

func mockConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				Key910: "ascend-910",
			},
		},
		Data: map[string]string{
			ConfigmapKey: "{\"status\":\"initializing\"}",
		},
		BinaryData: nil,
	}
}

// TestNewController test method NewContrller
func TestNewController(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockV1.EXPECT().Events(gomock.Any()).Return(nil).Times(three)
	mockV1.EXPECT().Pods(gomock.Any()).Return(nil)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(two)
	mockV1.Events("")
	mockInformer := mock_v1alpha1.NewMockJobInformer(ctrl)
	mockShared := mock_cache.NewMockSharedIndexInformer(ctrl)
	mockShared.EXPECT().AddEventHandler(gomock.Any()).Return()
	mockShared.EXPECT().GetIndexer().Return(nil)
	mockInformer.EXPECT().Informer().Return(mockShared).Times(three)
	stub := gostub.StubFunc(&newBusinessAgent, createAgentForController(false), nil)
	defer stub.Reset()
	tests := []struct {
		want *Controller
		name string
	}{
		{
			name: "normal situation,return controller instance",
			want: &Controller{
				businessAgent: createAgentForController(false),
			},
		},
	}
	config := &Config{
		DryRun:           false,
		DisplayStatistic: false,
		PodParallelism:   1,
		CmCheckInterval:  decimal,
		CmCheckTimeout:   oneMinitue,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewController(mockK8s, nil, config, mockInformer, make(chan struct{}))
			if !reflect.DeepEqual(got.businessAgent, tt.want.businessAgent) {
				t.Errorf("NewController() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestController_Run test run
func TestController_Run(t *testing.T) {
	type args struct {
		threadiness        int
		monitorPerformance bool
		stopCh             chan struct{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "normal situation,no error returned",
			wantErr: false,
			args: args{
				threadiness:        1,
				monitorPerformance: false,
				stopCh:             make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				kubeclientset: nil,
				jobclientset:  nil,
				jobsSynced: func() bool {
					return true
				},
				jobsIndexer:        nil,
				businessAgent:      createAgent(false),
				workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Jobs"),
				recorder:           nil,
				workAgentInterface: createAgent(false),
			}
			go func() {
				time.Sleep(1 * time.Second)
				tt.args.stopCh <- struct{}{}
			}()
			if err := c.Run(tt.args.threadiness, tt.args.monitorPerformance, tt.args.stopCh); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestController_enqueueJob  test enqueueJob
func TestController_enqueueJob(t *testing.T) {
	tests := []struct {
		name      string
		obj       interface{}
		eventType string
	}{
		{
			name:      "test1: Jod be added to queue",
			obj:       mockJob(),
			eventType: "add",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockController()
			c.enqueueJob(tt.obj, tt.eventType)
			job, _ := c.workqueue.Get()
			assert.NotNil(t, job)
		})

	}
}

func mockController() *Controller {
	return &Controller{
		kubeclientset:      nil,
		jobclientset:       nil,
		jobsSynced:         nil,
		jobsIndexer:        nil,
		businessAgent:      createAgent(false),
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Jobs"),
		recorder:           nil,
		workAgentInterface: createAgent(false),
	}
}

// TestController_processNextWorkItem  test  processNextWorkItem
func TestController_processNextWorkItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockIndexr := mock_cache.NewMockIndexer(ctrl)
	mockIndexr.EXPECT().GetByKey(gomock.Any()).Return(mockJob(), true, nil).Times(four)
	var controllers []*Controller
	contrl := mockController()
	contrl.workqueue.AddRateLimited("vcjob/testpod/delete")
	contrl.jobsIndexer = mockIndexr
	controllers = append(controllers, contrl)
	contrl2 := mockController()
	contrl2.workqueue.AddRateLimited(nil)
	controllers = append(controllers, contrl2)
	tests := []struct {
		controller *Controller
		name       string
		want       bool
	}{
		{
			name:       "test1: normal situation, retrun true",
			controller: controllers[0],
			want:       true,
		},
		{
			name:       "test2: obj format error, return true",
			controller: controllers[1],
			want:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.controller.processNextWorkItem(); got != tt.want {
				t.Errorf("processNextWorkItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
