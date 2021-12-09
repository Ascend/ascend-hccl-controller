/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package cmd implements initialization of the startup parameters of the hccl-controller
package main

import (
	"flag"
	"fmt"
	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/model"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"

	"hccl-controller/pkg/resource-controller/signals"
	"hccl-controller/pkg/ring-controller/controller"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	cinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	vkClientset "volcano.sh/apis/pkg/client/clientset/versioned"
	informers "volcano.sh/apis/pkg/client/informers/externalversions"
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
	cmCheckInterval    = 2
	cmCheckTimeout     = 10
	defaultLogFileName = "/var/log/mindx-dl/hccl-controller/hccl-controller.log"
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
	hwlog.RunLog.Infof("hccl controller starting and the version is %s", BuildVersion)
	if hcclVersion != "v1" && hcclVersion != "v2" {
		hwlog.RunLog.Fatalf("invalid json version value, should be v1/v2")
	}
	agent.JSONVersion = hcclVersion

	// check the validity of input parameters
	if jobParallelism <= 0 {
		hwlog.RunLog.Fatalf("Error parsing parameters: parallelism should be a positive integer.")
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Fatalf("Error building kubeconfig")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		hwlog.RunLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	jobClient, err := vkClientset.NewForConfig(cfg)
	if err != nil {
		hwlog.RunLog.Fatalf("Error building job clientset: %s", err.Error())
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
	if err = control.Run(jobParallelism, stopCh); err != nil {
		hwlog.RunLog.Fatalf("Error running controller: %s", err.Error())
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
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info(default), 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup operation log files, must be greater than or equal to 7 days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFileName,
		"Log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup log files, range (0, 30].")

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
	if err := hwlog.InitRunLogger(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}
}
