/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

const (
	// Key910 to get Configmap
	Key910 = "ring-controller.atlas"
	// Val910 to get Configmap
	Val910 = "ascend-910"
	// ConfigmapPrefix to get from configmap
	ConfigmapPrefix = "rings-config"
	// ConfigmapCompleted Staus
	ConfigmapCompleted = "completed"
	// ConfigmapInitializing status
	ConfigmapInitializing = "initializing"
	// ConfigmapKey configmap Data Name
	ConfigmapKey = "hccl.json"
	// PodDeviceKey Pod annoation Key
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankIndexKey pod rank index
	PodRankIndexKey = "hccl/rankIndex"
	// DeploymentNameKey pod label
	DeploymentNameKey = "deploy-name"
	// WorkloadLNameLabelKey job name label key for k8s job and meituan CRD: tfjob, medaljob, mpijob
	WorkloadLNameLabelKey = "job-name"
	// EventAdd event add
	EventAdd = "add"
	// EventUpdate event to update
	EventUpdate = "update"
	// EventDelete event to delete
	EventDelete = "delete"

	retryMilliSecond = 5
	threeMinutes     = 180
	splitNum         = 4
)

var (
	// JSONVersion of hccl.json
	JSONVersion = "v2"
	// ResourceList pod annotation
	ResourceList = []string{"huawei.com/Ascend910", "huawei.com/Ascend910-2c", "huawei.com/Ascend910-4c",
		"huawei.com/Ascend910-8c", "huawei.com/Ascend910-16c"}
)

// BusinessAgent Agent for all businessWorkers, responsibilities:
// * list/watch 910 pods, and assign each pod to corresponding handler
//   (each business worker belongs to a volcano job, and contains a handler for building rank table)
type BusinessAgent struct {
	// Config Agent configuration file
	Config *Config
	// business worker for each volcano job
	BusinessWorker  map[string]Worker
	informerFactory informers.SharedInformerFactory
	podInformer     cache.SharedIndexInformer
	// PodsIndexer to get pod index by namespace&name
	PodsIndexer cache.Indexer
	// KubeClientSet : ClientSet to contact kube apiServer
	KubeClientSet kubernetes.Interface
	agentSwitch   <-chan struct{}

	// RwMutex : to lock Agent Resource eg. Workqueue & BusinessWorker
	RwMutex sync.RWMutex

	// event recorder
	recorder record.EventRecorder
	// Workqueue: A queue with a limited rate.This queue is used to put pod event information
	Workqueue workqueue.RateLimitingInterface

	// if print only, do not delete anything.
	dryRun bool
}

// Config controller init configure
type Config struct {
	// DryRun:Is it a test
	DryRun bool
	// DisplayStatistic : a flag if starts to report rank table build statistic for job
	DisplayStatistic bool
	// PodParallelism : how many goroutine to run in the agent
	PodParallelism int
	// CmCheckInterval: ConfigMap Interval
	CmCheckInterval int
	// CmCheckTimeout :ConfigMap TimeOut
	CmCheckTimeout int
}

type podIdentifier struct {
	namespace string
	name      string
	jobName   string
	eventType string
}

// CommonPodWorker job pod worker for rank table configmap generation
type CommonPodWorker struct {
	CommonPodWorkerInfo
	CommonPodInfo
}

// CommonPodWorkerInfo ：normal Worker info
type CommonPodWorkerInfo struct {
	kubeclientset     kubernetes.Interface
	recorder          record.EventRecorder
	cmMu, statisticMu sync.Mutex
	dryRun            bool
	statisticSwitch   chan struct{}

	podsIndexer cache.Indexer

	configmapName string
	configmapData v1.RankTabler

	statisticStopped  bool
	rankIndex         int
	cachedPodNum      int32
	taskReplicasTotal int32
	labelKey          string
	labelVal          string
}

// CommonPodInfo ： common pod info
type CommonPodInfo struct {
	// CreationTimestamp is needed to distinguish cases between: 1. old pod + new worker  OR  2. new pod + old worker
	CreationTimestamp metav1.Time
	// Namespace :Job namespace
	Namespace string
	// Name : Job name
	Name string
}
