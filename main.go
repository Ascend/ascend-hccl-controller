/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package cmd implements initialization of the startup parameters of the hccl-controller
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/client/clientset/versioned"
	"volcano.sh/apis/pkg/client/informers/externalversions"

	"hccl-controller/pkg/resource-controller/signals"
	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
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
	// KubeConfig kubernetes config file
	KubeConfig string
)

const (
	dryRun             = false
	displayStatistic   = false
	cmCheckInterval    = 2
	cmCheckTimeout     = 10
	defaultLogFileName = "/var/log/mindx-dl/hccl-controller/hccl-controller.log"
	defaultKubeConfig  = "/etc/mindx-dl/hccl-controller/.config/config6"
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("HCCL-Controller version: %s \n", BuildVersion)
		os.Exit(0)
	}
	stopLogCh := make(chan struct{})
	defer close(stopLogCh)
	if err := initHwLogger(stopLogCh); err != nil {
		fmt.Printf("%v", err)
		return
	}
	hwlog.RunLog.Infof("hccl controller starting and the version is %s", BuildVersion)
	if hcclVersion != "v1" && hcclVersion != "v2" {
		hwlog.RunLog.Fatalf("invalid json version value, should be v1/v2")
	}
	agent.JSONVersion = hcclVersion
	validateParallelism()
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	if KubeConfig == "" && utils.IsExists(defaultKubeConfig) {
		KubeConfig = defaultKubeConfig
	}
	cfg, err := getK8sConfig()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		hwlog.RunLog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	jobClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		hwlog.RunLog.Fatalf("Error building job clientset: %s", err.Error())
	}

	jobInformerFactory, deploymentFactory := newInformerFactory(jobClient, kubeClient)
	jobInformer := jobInformerFactory.Batch().V1alpha1().Jobs()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.VCJobType] = jobInformer.Informer().GetIndexer()
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()

	control := controller.NewController(kubeClient, jobClient, newConfig(),
		controller.InformerInfo{JobInformer: jobInformer, DeployInformer: deploymentInformer,
			CacheIndexers: cacheIndexer}, stopCh)

	go jobInformerFactory.Start(stopCh)
	go deploymentFactory.Start(stopCh)
	if err = control.Run(jobParallelism, stopCh); err != nil {
		hwlog.RunLog.Fatalf("Error running controller: %s", err.Error())
	}
}

func newConfig() *agent.Config {
	config := &agent.Config{
		DryRun:           dryRun,
		DisplayStatistic: displayStatistic,
		PodParallelism:   podParallelism,
		CmCheckInterval:  cmCheckInterval,
		CmCheckTimeout:   cmCheckTimeout,
	}
	return config
}

func newInformerFactory(jobClient *versioned.Clientset, kubeClient *kubernetes.Clientset) (
	externalversions.SharedInformerFactory, informers.SharedInformerFactory) {
	labelSelector := labels.Set(map[string]string{agent.Key910: agent.Val910}).AsSelector().String()
	jobInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(jobClient,
		time.Second*common.InformerInterval, externalversions.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = labelSelector
		}))
	deploymentFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient,
		time.Second*common.InformerInterval, informers.WithTweakListOptions(func(options *v1.ListOptions) {
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
		"Parallelism of job events handling, it should be range [1, 32].")
	flag.IntVar(&podParallelism, "podParallelism", 1,
		"Parallelism of pod events handling, it should be range [1, 32].")
	flag.BoolVar(&version, "version", false,
		"Query the verison of the program")
	flag.StringVar(&hcclVersion, "json", "v2",
		"Select version of hccl json file (v1/v2).")
	flag.StringVar(&KubeConfig, "kubeConfig", "", "Path to a kubeconfig. "+
		"Only required if out-of-cluster.")

}

func initHwLogger(stopCh chan struct{}) error {
	if err := hwlog.InitRunLogger(hwLogConfig, stopCh); err != nil {
		return fmt.Errorf("hwlog init failed, error is %v", err)
	}

	return nil
}

func validateParallelism() {
	// check the validity of input parameters jobParallelism
	if jobParallelism <= 0 || jobParallelism > common.MaxJobParallelism {
		hwlog.RunLog.Fatalf("Error parsing parameters: job parallelism should be range [1, 32].")
	}
	// check the validity of input parameters podParallelism
	if podParallelism <= 0 || podParallelism > common.MaxPodParallelism {
		hwlog.RunLog.Fatalf("Error parsing parameters: pod parallelism should be range [1, 32].")
	}
}

func getK8sConfig() (*rest.Config, error) {
	path, err := utils.CheckPath(KubeConfig)
	if err != nil {
		return nil, err
	}
	cfg, err := utils.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, errors.New("error building kubeconfig")
	}

	return cfg, nil
}
