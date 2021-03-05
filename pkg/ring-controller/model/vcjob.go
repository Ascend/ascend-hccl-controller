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

package model

import (
	"fmt"
	agent2 "hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	v2 "hccl-controller/pkg/ring-controller/ranktable/v2"
	appsV1 "k8s.io/api/apps/v1"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"strconv"
	"time"
	v1alpha1apis "volcano.sh/volcano/pkg/apis/batch/v1alpha1"
)

// ResourceEventHandler to define same func, controller to use this function to finish some thing.
type ResourceEventHandler interface {
	EventAdd(tagentInterface *agent2.BusinessAgent) error
	EventUpdate(tagentInterface *agent2.BusinessAgent) error
	GenerateGrouplist() ([]*v1.Group, int32, error)
	GetReplicas() string
	GetCacheIndex() cache.Indexer
	GetModelKey() string
}

// GetModelKey: return model key.
func (model *modelCommon) GetModelKey() string {
	return model.key
}

// GetCacheIndex: return CacheIndex
func (model *modelCommon) GetCacheIndex() cache.Indexer {
	return model.cacheIndexer
}

// GetReplicas : return vcjob replicas
func (job *VCJobModel) GetReplicas() string {
	return strconv.Itoa(len(job.taskSpec))
}

// EventAdd to handle vcjob add event
func (job *VCJobModel) EventAdd(agent *agent2.BusinessAgent) error {

	agent.RwMutex.RLock()
	klog.V(L2).Infof("create business worker for %s/%s", job.JobNamespace, job.JobName)
	_, exist := agent.BusinessWorker[job.JobNamespace+"/"+job.JobName]
	if exist {
		agent.RwMutex.RUnlock()
		klog.V(L2).Infof("business worker for %s/%s is already existed", job.JobNamespace, job.JobName)
		return nil
	}
	agent.RwMutex.RUnlock()

	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(job.JobNamespace, job.JobName, agent.KubeClientSet, agent.Config)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString := cm.Data[agent2.ConfigmapKey]
	klog.V(L3).Info("jobstarting==>", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(job, jobStartString, agent2.JSONVersion)
	if err != nil {
		return err
	}
	jobWorker := agent2.NewVCJobWorker(agent, job.JobInfo, ranktable, replicasTotal)

	// create a business worker for current job
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()

	// start to report rank table build statistic for current job
	if agent.Config.DisplayStatistic {
		go jobWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	agent.BusinessWorker[job.JobNamespace+"/"+job.JobName] = jobWorker
	return nil
}

// EventUpdate : to handle vcjob update event
func (job *VCJobModel) EventUpdate(agent *agent2.BusinessAgent) error {
	if job.jobPhase == JobRestartPhase {
		agent2.DeleteWorker(job.JobNamespace, job.JobName, agent)
		return nil
	}
	agent.RwMutex.RLock()
	_, exist := agent.BusinessWorker[job.JobNamespace+"/"+job.JobName]
	agent.RwMutex.RUnlock()
	if !exist {
		// for job update, if create business worker at job restart phase, the version will be incorrect
		err := job.EventAdd(agent)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateGrouplist ： to generate GroupList, ranktable v1 will use it.
func (job *VCJobModel) GenerateGrouplist() ([]*v1.Group, int32, error) {
	var replicasTotal int32
	var groupList []*v1.Group
	for _, taskSpec := range job.taskSpec {
		var deviceTotal int32

		for _, container := range taskSpec.Template.Spec.Containers {
			quantity, exist := container.Resources.Limits[agent2.ResourceName]
			quantityValue := int32(quantity.Value())
			if exist && quantityValue > 0 {
				deviceTotal += quantityValue
			}
		}
		deviceTotal *= taskSpec.Replicas

		var instanceList []*v1.Instance
		group := v1.Group{GroupName: taskSpec.Name, DeviceCount: strconv.FormatInt(int64(deviceTotal), decimal),
			InstanceCount: strconv.FormatInt(int64(taskSpec.Replicas), decimal), InstanceList: instanceList}
		groupList = append(groupList, &group)
		replicasTotal += taskSpec.Replicas
	}
	return groupList, replicasTotal, nil
}

// checkCMCreation check configmap
func checkCMCreation(namespace, name string, kubeClientSet kubernetes.Interface, config *agent2.Config) (
	*apiCoreV1.ConfigMap, error) {
	var cm *apiCoreV1.ConfigMap
	err := wait.PollImmediate(time.Duration(config.CmCheckTimeout)*time.Second,
		time.Duration(config.CmCheckTimeout)*time.Second,
		func() (bool, error) {
			var errTmp error

			cm, errTmp = kubeClientSet.CoreV1().ConfigMaps(namespace).
				Get(fmt.Sprintf("%s-%s",
					agent2.ConfigmapPrefix, name), metav1.GetOptions{})
			if errTmp != nil {
				if errors.IsNotFound(errTmp) {
					return false, nil
				}
				return true, fmt.Errorf("get configmap error: %v", errTmp)
			}
			return true, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap for job %s/%s: %v", namespace, name, err)
	}
	label910, exist := (*cm).Labels[agent2.Key910]
	if !exist || (exist && label910 != agent2.Val910) {
		return nil, fmt.Errorf("invalid configmap label" + label910)
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
	switch t := obj.(type) {
	case *v1alpha1apis.Job:
		model = &VCJobModel{modelCommon: modelCommon{key: key, cacheIndexer: indexers[VCJobType]},
			JobInfo: agent2.JobInfo{JobUID: string(t.UID), JobVersion: t.Status.Version,
				JobCreationTimestamp: t.CreationTimestamp, JobNamespace: t.Namespace, JobName: t.Name},
			jobPhase: string(t.Status.State.Phase), taskSpec: t.Spec.Tasks}
	case *appsV1.Deployment:
		model = &DeployModel{modelCommon: modelCommon{key: key, cacheIndexer: indexers[DeploymentType]},
			containers: t.Spec.Template.Spec.Containers, replicas: *t.Spec.Replicas,
			DeployInfo: agent2.DeployInfo{DeployNamespace: t.Namespace, DeployName: t.Name,
				DeployCreationTimestamp: t.CreationTimestamp}}
	default:
		return nil, fmt.Errorf("job factory err, %s ", key)
	}

	return model, nil
}

// RanktableFactory : return the version type of ranktable according to your input parameters
func RanktableFactory(model ResourceEventHandler, jobStartString, JSONVersion string) (v1.RankTabler, int32, error) {
	var ranktable v1.RankTabler
	groupList, replicasTotal, err := model.GenerateGrouplist()
	if err != nil {
		return nil, 0, fmt.Errorf("generate group list from job error: %v", err)
	}

	if JSONVersion == "v1" {
		var configmapData v1.RankTable
		err = configmapData.UnmarshalToRankTable(jobStartString)
		if err != nil {
			return nil, 0, err
		}
		ranktable = &v1.RankTable{RankTableStatus: v1.RankTableStatus{Status: agent2.ConfigmapInitializing},
			GroupCount: model.GetReplicas(), GroupList: groupList}
	} else {
		var configmapData v2.RankTable
		err = configmapData.UnmarshalToRankTable(jobStartString)
		if err != nil {
			return nil, 0, err
		}
		var serverList []*v2.Server
		ranktable = &v2.RankTable{ServerCount: strconv.Itoa(len(serverList)), ServerList: serverList,
			RankTableStatus: v1.RankTableStatus{Status: agent2.ConfigmapInitializing}, Version: "1.0"}
	}
	return ranktable, replicasTotal, nil
}
