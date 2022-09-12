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
	"errors"
	"strconv"

	"hccl-controller/pkg/hwlog"
	agent2 "hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// EventAdd : to handle job add event
func (jm *CommonWorkload) EventAdd(agent *agent2.BusinessAgent) error {
	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(agent.KubeClientSet, agent.Config, jm.Namespace, jm.Name, jm.labelKey, jm.labelVal)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString, ok := cm.Data[agent2.ConfigmapKey]
	if !ok {
		return errors.New("The key of " + agent2.ConfigmapKey + "does not exist")
	}
	hwlog.Debug("jobstarting==>", jobStartString)

	ranktable, replicasTotal, err := RanktableFactory(jm, jobStartString, agent2.JSONVersion)
	if err != nil {
		return err
	}
	jobWorker := agent2.NewCommonPodWorker(agent, jm.CommonPodInfo, ranktable, replicasTotal, jm.labelKey, jm.labelVal)

	// create a business worker for current job
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()

	hwlog.Infof("create business worker for %s/%s", jm.Namespace, jm.Name)
	_, exist := agent.BusinessWorker[jm.Namespace+"/"+jm.Name]
	if exist {
		hwlog.Infof("business worker for %s/%s is already existed", jm.Namespace, jm.Name)
		return nil
	}

	// start to report rank table build statistic for current job
	if agent.Config.DisplayStatistic {
		go jobWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	agent.BusinessWorker[jm.Namespace+"/"+jm.Name] = jobWorker
	return nil
}

// EventUpdate : to handle job update event
func (jm *CommonWorkload) EventUpdate(agent *agent2.BusinessAgent) error {
	agent.RwMutex.RLock()
	_, exist := agent.BusinessWorker[jm.Namespace+"/"+jm.Name]
	agent.RwMutex.RUnlock()
	if !exist {
		// for pod update,  the version will be incorrect
		err := jm.EventAdd(agent)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateGrouplist ： to generate GroupList, ranktable v1 will use it.
func (jm *CommonWorkload) GenerateGrouplist() ([]*v1.Group, int32, error) {
	var groupList []*v1.Group
	var deviceTotal int32

	for _, container := range jm.containers {
		deviceTotal += agent2.GetNPUNum(container)
	}
	deviceTotal *= jm.replicas

	var instanceList []*v1.Instance
	group := v1.Group{
		GroupName:     jm.Name,
		DeviceCount:   strconv.FormatInt(int64(deviceTotal), decimal),
		InstanceCount: strconv.FormatInt(int64(jm.replicas), decimal),
		InstanceList:  instanceList}
	groupList = append(groupList, &group)

	return groupList, jm.replicas, nil
}

// GetReplicas : return job replicas
func (jm *CommonWorkload) GetReplicas() string {
	return strconv.Itoa(int(jm.replicas))
}
