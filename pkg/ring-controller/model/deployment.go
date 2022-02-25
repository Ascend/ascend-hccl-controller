/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package model : to handle event in controller logic
package model

import (
	"errors"
	agent2 "hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	"huawei.com/npu-exporter/hwlog"
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
	jobStartString, ok := cm.Data[agent2.ConfigmapKey]
	if !ok {
		return errors.New("the key of " + agent2.ConfigmapKey + " does not exist")
	}
	var rst v1.RankTableStatus
	if err = rst.UnmarshalToRankTable(jobStartString); err != nil {
		return err
	}
	hwlog.RunLog.Debug("jobStarting: ", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(deploy, rst, agent2.JSONVersion)
	if err != nil {
		return err
	}
	deploymentWorker := agent2.NewDeploymentWorker(agent, deploy.DeployInfo, ranktable, replicasTotal)

	// create a business worker for current deployment
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()

	hwlog.RunLog.Infof("create business worker for %s/%s", deploy.DeployNamespace, deploy.DeployName)
	_, exist := agent.BusinessWorker[deploy.DeployNamespace+"/"+deploy.DeployName]
	if exist {
		hwlog.RunLog.Infof("business worker for %s/%s is already existed", deploy.DeployNamespace, deploy.DeployName)
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
		deviceTotal += agent2.GetNPUNum(container)
	}
	deviceTotal *= deploy.replicas

	var instanceList []*v1.Instance
	group := v1.Group{GroupName: deploy.DeployName, DeviceCount: strconv.FormatInt(int64(deviceTotal),
		common.Decimal), InstanceCount: strconv.FormatInt(int64(deploy.replicas), common.Decimal),
		InstanceList: instanceList}
	groupList = append(groupList, &group)

	return groupList, deploy.replicas, nil
}
