/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"huawei.com/npu-exporter/hwlog"
	appsV1 "k8s.io/api/apps/v1"
	apiCoreV1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	v1alpha1apis "volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	ranktablev2 "hccl-controller/pkg/ring-controller/ranktable/v2"
)

// ResourceEventHandler to define same func, controller to use this function to finish some thing.
type ResourceEventHandler interface {
	EventAdd(tagentInterface *agent.BusinessAgent) error
	EventUpdate(tagentInterface *agent.BusinessAgent) error
	GenerateGrouplist() ([]*ranktablev1.Group, int32, error)
	GetReplicas() string
	GetCacheIndex() cache.Indexer
	GetModelKey() string
}

// GetModelKey return model key.
func (model *modelCommon) GetModelKey() string {
	return model.key
}

// GetCacheIndex return CacheIndex
func (model *modelCommon) GetCacheIndex() cache.Indexer {
	return model.cacheIndexer
}

// GetReplicas : return vcjob replicas
func (job *VCJobModel) GetReplicas() string {
	return strconv.Itoa(len(job.taskSpec))
}

// EventAdd to handle vcjob add event
func (job *VCJobModel) EventAdd(businessAgent *agent.BusinessAgent) error {

	businessAgent.RwMutex.RLock()
	hwlog.RunLog.Infof("create business worker for %s/%s", job.JobNamespace, job.JobName)
	_, exist := businessAgent.BusinessWorker[job.JobNamespace+"/"+job.JobName]
	businessAgent.RwMutex.RUnlock()
	if exist {
		hwlog.RunLog.Infof("business worker for %s/%s is already existed", job.JobNamespace, job.JobName)
		return nil
	}

	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(job.JobNamespace, job.JobName, businessAgent.KubeClientSet, businessAgent.Config)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString, ok := cm.Data[agent.ConfigmapKey]
	if !ok {
		return errors.New("the key of " + agent.ConfigmapKey + " does not exist")
	}
	var rst ranktablev1.RankTableStatus
	if err = rst.UnmarshalToRankTable(jobStartString); err != nil {
		return err
	}
	hwlog.RunLog.Info("jobStarting: ", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(job, rst, agent.GetJSONVersion())
	if err != nil {
		return err
	}
	jobWorker := agent.NewVCJobWorker(businessAgent, job.JobInfo, ranktable, replicasTotal)

	// create a business worker for current job
	businessAgent.RwMutex.Lock()
	defer businessAgent.RwMutex.Unlock()

	// start to report rank table build statistic for current job
	if businessAgent.Config.DisplayStatistic {
		go jobWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	businessAgent.BusinessWorker[job.JobNamespace+"/"+job.JobName] = jobWorker
	return nil
}

// EventUpdate : to handle vcjob update event
func (job *VCJobModel) EventUpdate(businessAgent *agent.BusinessAgent) error {
	businessAgent.RwMutex.RLock()
	_, exist := businessAgent.BusinessWorker[job.JobNamespace+"/"+job.JobName]
	businessAgent.RwMutex.RUnlock()
	if !exist {
		// for job update, if create business worker at job restart phase, the version will be incorrect
		err := job.EventAdd(businessAgent)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateGrouplist ： to generate GroupList, ranktable v1 will use it.
func (job *VCJobModel) GenerateGrouplist() ([]*ranktablev1.Group, int32, error) {
	var replicasTotal int32
	var groupList []*ranktablev1.Group
	for _, taskSpec := range job.taskSpec {
		var deviceTotal int32

		for _, container := range taskSpec.Template.Spec.Containers {
			deviceTotal += agent.GetNPUNum(container)
		}
		deviceTotal *= taskSpec.Replicas

		var instanceList []*ranktablev1.Instance
		group := ranktablev1.Group{GroupName: taskSpec.Name, DeviceCount: strconv.FormatInt(int64(deviceTotal),
			common.Decimal), InstanceCount: strconv.FormatInt(int64(taskSpec.Replicas), common.Decimal),
			InstanceList: instanceList}
		groupList = append(groupList, &group)
		replicasTotal += taskSpec.Replicas
	}
	return groupList, replicasTotal, nil
}

// checkCMCreation check configmap
func checkCMCreation(namespace, name string, kubeClientSet kubernetes.Interface, config *agent.Config) (
	*apiCoreV1.ConfigMap, error) {
	var cm *apiCoreV1.ConfigMap
	err := wait.PollImmediate(time.Duration(config.CmCheckTimeout)*time.Second,
		time.Duration(config.CmCheckTimeout)*time.Second,
		func() (bool, error) {
			var errTmp error
			cm, errTmp = kubeClientSet.CoreV1().ConfigMaps(namespace).
				Get(context.TODO(), fmt.Sprintf("%s-%s", agent.ConfigmapPrefix, name), metav1.GetOptions{})
			if errTmp != nil {
				if apierrors.IsNotFound(errTmp) {
					return false, nil
				}
				return true, fmt.Errorf("get configmap error: %#v", errTmp)
			}
			return true, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap for job %s/%s: %v", namespace, name, err)
	}
	label910, exist := (*cm).Labels[agent.Key910]
	if !exist || label910 != agent.Val910 {
		return nil, fmt.Errorf("invalid configmap label %#v", label910)
	}

	return cm, nil
}

// Factory : to generate model
func Factory(obj interface{}, eventType string, indexers map[string]cache.Indexer) (ResourceEventHandler, error) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("object has no meta: %v", err)
	}
	key := metaData.GetName() + "/" + eventType
	if len(metaData.GetNamespace()) > 0 {
		key = metaData.GetNamespace() + "/" + metaData.GetName() + "/" + eventType
	}
	var model ResourceEventHandler
	if _, ok := indexers[VCJobType]; !ok {
		return nil, fmt.Errorf("the key does not exist err %v ", ok)
	}
	if _, ok := indexers[DeploymentType]; !ok {
		return nil, fmt.Errorf("the key does not exist err %v ", ok)
	}
	switch t := obj.(type) {
	case *v1alpha1apis.Job:
		model = &VCJobModel{modelCommon: modelCommon{key: key, cacheIndexer: indexers[VCJobType]},
			JobInfo: agent.JobInfo{JobUID: string(t.UID), JobVersion: t.Status.Version,
				JobCreationTimestamp: t.CreationTimestamp, JobNamespace: t.Namespace, JobName: t.Name},
			jobPhase: string(t.Status.State.Phase), taskSpec: t.Spec.Tasks}
	case *appsV1.Deployment:
		model = &DeployModel{modelCommon: modelCommon{key: key, cacheIndexer: indexers[DeploymentType]},
			containers: t.Spec.Template.Spec.Containers, replicas: *t.Spec.Replicas,
			DeployInfo: agent.DeployInfo{DeployNamespace: t.Namespace, DeployName: t.Name,
				DeployCreationTimestamp: t.CreationTimestamp}}
	default:
		return nil, fmt.Errorf("job factory err, %s ", key)
	}

	return model, nil
}

// RanktableFactory : return the version type of ranktable according to your input parameters
func RanktableFactory(model ResourceEventHandler, rst ranktablev1.RankTableStatus,
	JSONVersion string) (ranktablev1.RankTabler, int32, error) {
	var ranktable ranktablev1.RankTabler
	groupList, replicasTotal, err := model.GenerateGrouplist()
	if err != nil {
		return nil, 0, fmt.Errorf("generate group list from job error: %v", err)
	}
	if JSONVersion == "v1" {
		ranktable = &ranktablev1.RankTable{RankTableStatus: ranktablev1.RankTableStatus{Status: rst.Status},
			GroupCount: model.GetReplicas(), GroupList: groupList}
	} else {
		ranktable = &ranktablev2.RankTable{ServerCount: "0", ServerList: []*ranktablev2.Server(nil),
			RankTableStatus: ranktablev1.RankTableStatus{Status: rst.Status}, Version: "1.0"}
	}
	return ranktable, replicasTotal, nil
}
