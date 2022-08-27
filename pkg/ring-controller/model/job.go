package model

import (
	"errors"
	"hccl-controller/pkg/hwlog"
	agent2 "hccl-controller/pkg/ring-controller/agent"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	"strconv"
)

// EventAdd : to handle job add event
func (jm *K8sJobModel) EventAdd(agent *agent2.BusinessAgent) error {
	// check if job's corresponding configmap is created successfully via volcano controller
	cm, err := checkCMCreation(jm.JobNamespace, jm.JobName, agent.KubeClientSet, agent.Config)
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
	jobWorker := agent2.NewK8SJobWorker(agent, jm.K8sJobInfo, ranktable, replicasTotal)

	// create a business worker for current job
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()

	hwlog.Infof("create business worker for %s/%s", jm.JobNamespace, jm.JobName)
	_, exist := agent.BusinessWorker[jm.JobNamespace+"/"+jm.JobName]
	if exist {
		hwlog.Infof("business worker for %s/%s is already existed", jm.JobNamespace, jm.JobName)
		return nil
	}

	// start to report rank table build statistic for current job
	if agent.Config.DisplayStatistic {
		go jobWorker.Statistic(BuildStatInterval)
	}

	// save current business worker
	agent.BusinessWorker[jm.JobNamespace+"/"+jm.JobName] = jobWorker
	return nil
}

// EventUpdate : to handle job update event
func (jm *K8sJobModel) EventUpdate(agent *agent2.BusinessAgent) error {
	agent.RwMutex.RLock()
	_, exist := agent.BusinessWorker[jm.JobNamespace+"/"+jm.JobName]
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
func (jm *K8sJobModel) GenerateGrouplist() ([]*v1.Group, int32, error) {
	var groupList []*v1.Group
	var deviceTotal int32

	for _, container := range jm.containers {
		deviceTotal += agent2.GetNPUNum(container)
	}
	deviceTotal *= jm.replicas

	var instanceList []*v1.Instance
	group := v1.Group{
		GroupName:     jm.JobName,
		DeviceCount:   strconv.FormatInt(int64(deviceTotal), decimal),
		InstanceCount: strconv.FormatInt(int64(jm.replicas), decimal),
		InstanceList:  instanceList}
	groupList = append(groupList, &group)

	return groupList, jm.replicas, nil
}

// GetReplicas : return job replicas
func (jm *K8sJobModel) GetReplicas() string {
	return strconv.Itoa(int(jm.replicas))
}
