/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package model : to handle event in controller logic
package model

import (
	"errors"
	"fmt"
	"strconv"

	"huawei.com/npu-exporter/hwlog"

	"hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
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
	var rst ranktablev1.RankTableStatus
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
func (deploy *DeployModel) GenerateGrouplist() ([]*ranktablev1.Group, int32, error) {
	var groupList []*ranktablev1.Group
	var deviceTotal int32

	for _, container := range deploy.containers {
		npuNum := agent.GetNPUNum(container)
		if npuNum == agent.InvalidNPUNum {
			return nil, 0, fmt.Errorf("get wrong npu num(%d) in container", npuNum)
		}
		deviceTotal += npuNum
	}
	deviceTotal *= deploy.replicas

	var instanceList []*ranktablev1.Instance
	group := ranktablev1.Group{GroupName: deploy.DeployName, DeviceCount: strconv.FormatInt(int64(deviceTotal),
		common.Decimal), InstanceCount: strconv.FormatInt(int64(deploy.replicas), common.Decimal),
		InstanceList: instanceList}
	groupList = append(groupList, &group)

	return groupList, deploy.replicas, nil
}
