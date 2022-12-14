/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package model : to handle event in controller logic
package model

import (
	"errors"
	"fmt"
	"strconv"

	"huawei.com/mindx/common/hwlog"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	"hccl-controller/pkg/ring-controller/ranktable/v1"
)

// GetReplicas : to return the replicas in deployment.
func (deploy *DeployModel) GetReplicas() string {
	return strconv.Itoa(int(deploy.replicas))
}

// EventAdd : to handle deployment add event
func (deploy *DeployModel) EventAdd(businessAgent *agent.BusinessAgent) error {
	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(deploy.DeployNamespace, deploy.DeployName, businessAgent.KubeClientSet,
		businessAgent.Config)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString, ok := cm.Data[agent.ConfigmapKey]
	if !ok {
		return errors.New("the key of " + agent.ConfigmapKey + " does not exist")
	}
	var rst v1.RankTableStatus
	if err = rst.UnmarshalToRankTable(jobStartString); err != nil {
		return err
	}
	hwlog.RunLog.Debugf("jobStarting: %#v", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(deploy, rst, agent.GetJSONVersion())
	if err != nil {
		return err
	}
	deploymentWorker := agent.NewDeploymentWorker(businessAgent, deploy.DeployInfo, ranktable, replicasTotal)

	// create a business worker for current deployment
	businessAgent.RwMutex.Lock()
	defer businessAgent.RwMutex.Unlock()

	hwlog.RunLog.Infof("create business worker for %s/%s", deploy.DeployNamespace, deploy.DeployName)
	_, exist := businessAgent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName]
	if exist {
		hwlog.RunLog.Infof("business worker for %s/%s is already existed", deploy.DeployNamespace, deploy.DeployName)
		return nil
	}

	// start to report rank table build statistic for current deployment
	if businessAgent.Config.DisplayStatistic {
		go deploymentWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	businessAgent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName] = deploymentWorker
	return nil
}

// EventUpdate : to handle deployment update event
func (deploy *DeployModel) EventUpdate(businessAgent *agent.BusinessAgent) error {
	businessAgent.RwMutex.RLock()
	_, exist := businessAgent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName]
	businessAgent.RwMutex.RUnlock()
	if !exist {
		// for pod update,  the version will be incorrect
		err := deploy.EventAdd(businessAgent)
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
		npuNum := agent.GetNPUNum(container)
		if npuNum == agent.InvalidNPUNum {
			return nil, 0, fmt.Errorf("get wrong npu num(%d) in container", npuNum)
		}
		deviceTotal += npuNum
	}
	if deploy.replicas > maxNodeNum {
		return nil, 0, errors.New("the number of Replicas in a deployment is too large")
	}
	deviceTotal *= deploy.replicas

	var instanceList []*v1.Instance
	group := v1.Group{GroupName: deploy.DeployName, DeviceCount: strconv.FormatInt(int64(deviceTotal),
		common.Decimal), InstanceCount: strconv.FormatInt(int64(deploy.replicas), common.Decimal),
		InstanceList: instanceList}
	groupList = append(groupList, &group)

	return groupList, deploy.replicas, nil
}
