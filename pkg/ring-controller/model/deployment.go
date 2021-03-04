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

// Package model : to handle event in controller logic
package model

import (
	agent2 "hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	"k8s.io/klog"
	"strconv"
)

// GetReplicas : to return the replicas in deployment.
func (deploy *DeployModel) GetReplicas() string {
	return strconv.Itoa(int(deploy.replicas))
}

// EventAdd : to handle deployment add event
func (deploy *DeployModel) EventAdd(agent *agent2.BusinessAgent) error {
	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(deploy.DeployNamespace, deploy.DeployName, agent.KubeClientSet, agent.Config)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString := cm.Data[agent2.ConfigmapKey]
	klog.V(L4).Info("jobstarting==>", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(deploy, jobStartString, agent2.JSONVersion)
	if err != nil {
		return err
	}
	deploymentWorker := agent2.NewDeploymentWorker(agent, deploy.DeployInfo, ranktable, replicasTotal)

	// create a business worker for current deployment
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()

	klog.V(L2).Infof("create business worker for %s/%s", deploy.DeployNamespace, deploy.DeployName)
	_, exist := agent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName]
	if exist {
		klog.V(L2).Infof("business worker for %s/%s is already existed", deploy.DeployNamespace, deploy.DeployName)
		return nil
	}

	// start to report rank table build statistic for current deployment
	if agent.Config.DisplayStatistic {
		go deploymentWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	agent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName] = deploymentWorker
	return nil
}

// EventUpdate : to handle deployment update event
func (deploy *DeployModel) EventUpdate(agent *agent2.BusinessAgent) error {
	agent.RwMutex.RLock()
	_, exist := agent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName]
	agent.RwMutex.RUnlock()
	if !exist {
		// for pod update,  the version will be incorrect
		err := deploy.EventAdd(agent)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateGrouplist to create GroupList. in ranktable v1 will use it.
func (deploy *DeployModel) GenerateGrouplist() ([]*v1.Group, int32, error) {
	var groupList []*v1.Group
	var deviceTotal int32

	for _, container := range deploy.containers {
		quantity, exist := container.Resources.Limits[agent2.ResourceName]
		quantityValue := int32(quantity.Value())
		if exist && quantityValue > 0 {
			deviceTotal += quantityValue
		}
	}
	deviceTotal *= deploy.replicas

	var instanceList []*v1.Instance
	group := v1.Group{GroupName: deploy.DeployName, DeviceCount: strconv.FormatInt(int64(deviceTotal), decimal),
		InstanceCount: strconv.FormatInt(int64(deploy.replicas), decimal), InstanceList: instanceList}
	groupList = append(groupList, &group)

	return groupList, deploy.replicas, nil
}
