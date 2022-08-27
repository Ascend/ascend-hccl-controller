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
	"os"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	cinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"hccl-controller/pkg/hwlog"
	"hccl-controller/pkg/resource-controller/signals"
	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/controller"
	"hccl-controller/pkg/ring-controller/model"
)

var (
	jobParallelism int
	podParallelism int
	version        bool
	hcclVersion    string
	// BuildVersion  build version
	BuildVersion string
	hwLogConfig  = &hwlog.LogConfig{LogFileName: defaultLogFileName}
)

const (
	dryRun             = false
	displayStatistic   = false
	kubeEnvMaxLength   = 253
	cmCheckInterval    = 2
	cmCheckTimeout     = 10
	rootUID            = 0
	rootGID            = 0
	defaultLogFileName = "/var/log/mindx-dl/hccl-controller/hccl-controller.log"
	kubeCfgEnvKey      = "KUBECONFIG"
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("HCCL-Controller version: %s \n", BuildVersion)
		os.Exit(0)
	}

	stopLogCh := make(chan struct{})
	defer close(stopLogCh)
	initHwLogger(stopLogCh)

	if hcclVersion != "v1" && hcclVersion != "v2" {
		hwlog.Fatalf("invalid json version value, should be v1/v2")
	}
	agent.JSONVersion = hcclVersion

	// check the validity of input parameters
	if jobParallelism <= 0 {
		hwlog.Fatalf("Error parsing parameters: parallelism should be a positive integer.")
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfgPath := getK8SCfg()
	cfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		hwlog.Fatalf("Error building kubeconfig")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		hwlog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	deploymentFactory := newInformerFactory(kubeClient)
	config := newConifg()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	k8sJobInformer := deploymentFactory.Batch().V1().Jobs()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()
	cacheIndexer[model.K8sJobType] = k8sJobInformer.Informer().GetIndexer()

	control := controller.NewController(
		kubeClient,
		config,
		controller.InformerInfo{
			DeployInformer: deploymentInformer,
			K8sJobInformer: k8sJobInformer,
			CacheIndexers:  cacheIndexer},
		stopCh)

	go deploymentFactory.Start(stopCh)
	if err = control.Run(jobParallelism, stopCh); err != nil {
		hwlog.Fatalf("Error running controller: %s", err.Error())
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

func newInformerFactory(kubeClient *kubernetes.Clientset) cinformers.SharedInformerFactory {
	labelSelector := labels.Set(map[string]string{agent.Key910: agent.Val910}).AsSelector().String()
	deploymentFactory := cinformers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30,
		cinformers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = labelSelector
		}))
	return deploymentFactory
}

func init() {
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info(default), 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files")
	flag.BoolVar(&hwLogConfig.IsCompress, "isCompress", false,
		"Whether backup files need to be compressed (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFileName, "Log file path")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups, "Maximum number of backup log files")

	flag.IntVar(&jobParallelism, "jobParallelism", 1,
		"Parallelism of job events handling.")
	flag.IntVar(&podParallelism, "podParallelism", 1,
		"Parallelism of pod events handling.")
	flag.BoolVar(&version, "version", false,
		"Query the verison of the program")
	flag.StringVar(&hcclVersion, "json", "v2",
		"Select version of hccl json file (v1/v2).")

}

func initHwLogger(stopCh chan struct{}) {
	if err := hwlog.Init(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}
}

func getK8SCfg() string {
	kubeConfig := os.Getenv(kubeCfgEnvKey)
	if err := checkKubeConfig(kubeConfig); err != nil {
		hwlog.Errorf("check kube config failed: %v", err)
		return ""
	}
	return kubeConfig
}

func checkKubeConfig(kubeConfig string) error {
	if len(kubeConfig) > kubeEnvMaxLength {
		return fmt.Errorf("kube config length %d is bigger than %d", len(kubeConfig), kubeEnvMaxLength)
	}
	kubeConfigPathInfo, err := os.Stat(kubeConfig)
	if err != nil || os.IsNotExist(err) {
		return nil
	}
	stat, ok := kubeConfigPathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != rootUID || stat.Gid != rootGID {
		return fmt.Errorf("non-root owner group of the path")
	}
	return nil
}
