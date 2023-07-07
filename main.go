/* Copyright(C) 2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package cmd implements initialization of the startup parameters of the hccl-controller
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"huawei.com/npu-exporter/v3/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
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
		return
	}
	if err := initHwLogger(); err != nil {
		fmt.Printf("%v", err)
		return
	}
	hwlog.RunLog.Infof("hccl controller starting and the version is %s", BuildVersion)
	if err := validate(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	kubeClient, jobClient, err := NewClientK8s()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	jobInformerFactory, deploymentFactory, newErr := newInformerFactory(jobClient, kubeClient)
	if newErr != nil {
		hwlog.RunLog.Error(newErr)
		return
	}
	jobInformer := jobInformerFactory.Batch().V1alpha1().Jobs()
	deploymentInformer := deploymentFactory.Apps().V1().Deployments()
	cacheIndexer := make(map[string]cache.Indexer, 1)
	cacheIndexer[model.VCJobType] = jobInformer.Informer().GetIndexer()
	cacheIndexer[model.DeploymentType] = deploymentInformer.Informer().GetIndexer()
	control, err := controller.NewEventController(kubeClient, jobClient, newConfig(),
		controller.InformerInfo{JobInformer: jobInformer, DeployInformer: deploymentInformer,
			CacheIndexers: cacheIndexer}, stopCh)
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	go jobInformerFactory.Start(stopCh)
	go deploymentFactory.Start(stopCh)
	if err = control.Run(jobParallelism, stopCh); err != nil {
		hwlog.RunLog.Errorf("Error running controller: %s", err.Error())
	}
}

// NewClientK8s create k8s client
func NewClientK8s() (*kubernetes.Clientset, *versioned.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build client config err: %#v", err)
		return nil, nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}
	jobClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building job clientset: %s", err.Error())
	}
	return kubeClient, jobClient, nil

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
	externalversions.SharedInformerFactory, informers.SharedInformerFactory, error) {
	temp, newErr := labels.NewRequirement(agent.Key910, selection.In, []string{agent.Val910B, agent.Val910})
	if newErr != nil {
		hwlog.RunLog.Infof("newInformerFactory %s", newErr)
		return nil, nil, newErr
	}
	labelSelector := temp.String()
	jobInformerFactory := externalversions.NewSharedInformerFactoryWithOptions(jobClient,
		time.Second*common.InformerInterval, externalversions.WithTweakListOptions(func(options *v1.
			ListOptions) {
			options.LabelSelector = labelSelector
		}))
	deploymentFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient,
		time.Second*common.InformerInterval, informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = labelSelector
		}))
	return jobInformerFactory, deploymentFactory, nil
}

func init() {
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup operation log files, range [7, 700] days")
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
}

func initHwLogger() error {
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		return fmt.Errorf("hwlog init failed, error is %v\n", err)
	}
	return nil
}

func validate() error {
	if hcclVersion != "v1" && hcclVersion != "v2" {
		return errors.New("invalid json version value, should be v1/v2")
	}
	agent.SetJSONVersion(hcclVersion)
	// check the validity of input parameters jobParallelism
	if jobParallelism <= 0 || jobParallelism > common.MaxJobParallelism {
		return errors.New("error parsing parameters: job parallelism should be range [1, 32]")
	}
	// check the validity of input parameters podParallelism
	if podParallelism <= 0 || podParallelism > common.MaxPodParallelism {
		return errors.New("error parsing parameters: pod parallelism should be range [1, 32]")
	}
	return nil
}
