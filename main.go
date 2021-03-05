/*
* Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
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

// Package cmd implements initialization of the startup parameters of the hccl-controller
package main

import (
	"flag"
	"fmt"
	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/model"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"os"
	"path/filepath"
	"time"

	"hccl-controller/pkg/resource-controller/signals"
	"hccl-controller/pkg/ring-controller/controller"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	cinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	vkClientset "volcano.sh/volcano/pkg/client/clientset/versioned"
	informers "volcano.sh/volcano/pkg/client/informers/externalversions"
)

const (
	cmCheckIntervalConst = 2
	cmCheckTimeoutConst  = 10
)

var (
	masterIUrl         string
	kubeconfig         string
	dryRun             bool
	displayStatistic   bool
	monitorPerformance bool
	jobParallelism     int
	podParallelism     int
	cmCheckInterval    int
	cmCheckTimeout     int
	version            bool
	jsonVersion        string
	// BuildVersion  build version
	BuildVersion string
)

func validate(masterIUrl *string) bool {
	if *masterIUrl == "" {
		return true
	}
	realPath, err := filepath.Abs(*masterIUrl)
	if err != nil {
		klog.Fatalf("It's error when converted to an absolute path.")
		return false
	}
	masterIUrl = &realPath
	return true
}

func main() {
	flag.Parse()
	if !validate(&masterIUrl) {
		klog.Fatalf("file not in security directory")
	}
	if !validate(&kubeconfig) {
		klog.Fatalf("file not in security directory")
	}
	if jsonVersion != "v1" && jsonVersion != "v2" {
		klog.Fatalf("invalid json version value, should be v1/v2")
	}
	agent.JSONVersion = jsonVersion

	if version {
		fmt.Printf("HCCL-Controller version: %s \n", BuildVersion)
		os.Exit(0)
	}

	// check the validity of input parameters
	if jobParallelism <= 0 {
		klog.Fatalf("Error parsing parameters: parallelism should be a positive integer.")
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterIUrl, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	jobClient, err := vkClientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building job clientset: %s", err.Error())
	}

	jobInformerFactory, deploymentFactory := newInformerFactory(jobClient, kubeClient)
	config := newConifg()
	jobInformer := jobInformerFactory.Batch().V1alpha1().Jobs()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.VCJobType] = jobInformer.Informer().GetIndexer()
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()

	control := controller.NewController(kubeClient, jobClient, config, controller.InformerInfo{JobInformer: jobInformer,
		DeployInformer: deploymentInformer, CacheIndexers: cacheIndexer}, stopCh)

	go jobInformerFactory.Start(stopCh)
	go deploymentFactory.Start(stopCh)
	if err = control.Run(jobParallelism, monitorPerformance, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func newConifg() *agent.Config {
	config := &agent.Config{
		DryRun:           dryRun,
		DisplayStatistic: displayStatistic,
		PodParallelism:   podParallelism,
		CmCheckInterval:  cmCheckInterval,
		CmCheckTimeout:   cmCheckTimeout,
	}
	return config
}

func newInformerFactory(jobClient *vkClientset.Clientset, kubeClient *kubernetes.Clientset) (
	informers.SharedInformerFactory, cinformers.SharedInformerFactory) {
	labelSelector := labels.Set(map[string]string{agent.Key910: agent.Val910}).AsSelector().String()
	jobInformerFactory := informers.NewSharedInformerFactoryWithOptions(jobClient, time.Second*30,
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = labelSelector
		}))
	deploymentFactory := cinformers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30,
		cinformers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = labelSelector
		}))
	return jobInformerFactory, deploymentFactory
}

func init() {
	// * buildStatInterval
	// * bdefaultResync of two informers
	// * bperiod of two runMasterWorker method
	klog.InitFlags(nil)

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterIUrl, "master", "",
		"The address of the Kubernetes API server. "+
			"Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.BoolVar(&dryRun, "dryRun", false,
		"Print only, do not delete anything.")
	flag.BoolVar(&displayStatistic, "displayStatistic", false,
		"Display progress of configmap updating.")
	flag.BoolVar(&monitorPerformance, "monitorPerformance", false,
		"Monitor performance of ring-controller.")

	flag.IntVar(&jobParallelism, "jobParallelism", 1,
		"Parallelism of job events handling.")
	flag.IntVar(&podParallelism, "podParallelism", 1,
		"Parallelism of pod events handling.")
	flag.IntVar(&cmCheckInterval, "cmCheckInterval", cmCheckIntervalConst,
		"Interval (seconds) to check job's configmap before building rank table.")
	flag.IntVar(&cmCheckTimeout, "cmCheckTimeout", cmCheckTimeoutConst,
		"Maximum time (seconds) to check creation of job's configmap.")
	flag.BoolVar(&version, "version", false,
		"Query the verison of the program")
	flag.StringVar(&jsonVersion, "json", "v2",
		"Select version of hccl json file (v1/v2).")
}
