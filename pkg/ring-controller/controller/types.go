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

// Package controller for controller
package controller

import (
	"hccl-controller/pkg/ring-controller/agent"
	v1 "k8s.io/client-go/informers/apps/v1"
	bv1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	clientset "volcano.sh/volcano/pkg/client/clientset/versioned"
	v1alpha1informers "volcano.sh/volcano/pkg/client/informers/externalversions/batch/v1alpha1"
)

const (
	controllerName = "ring-controller"
)

// Controller initialize business agent
type Controller struct {
	// component for recycle resources
	agent *agent.BusinessAgent

	cacheIndexers map[string]cache.Indexer
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// jobclientset is a clientset for volcano job
	jobclientset clientset.Interface

	// component for resource batch/v1alpha1/Job
	jobsSynced   cache.InformerSynced
	deploySynced cache.InformerSynced
	k8sJobSynced cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// InformerInfo : Defining what the Controller will use
type InformerInfo struct {
	// CacheIndexers : to store different type cache index
	CacheIndexers map[string]cache.Indexer
	// JobInformer : vcjob type informer
	JobInformer v1alpha1informers.JobInformer
	// DeployInformer: deployment type informer
	DeployInformer v1.DeploymentInformer
	//K8sJobInformer: job type informer
	K8sJobInformer bv1.JobInformer
}
