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

package agent

import (
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"

	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// NewDeploymentWorker ： to create Deployment Worker
func NewDeploymentWorker(agent *BusinessAgent, deploy DeployInfo, ranktable ranktablev1.RankTabler,
	replicasTotal int32) *DeployWorker {
	return &DeployWorker{WorkerInfo: WorkerInfo{kubeclientset: agent.KubeClientSet, podsIndexer: agent.PodsIndexer,
		recorder: agent.recorder, dryRun: agent.dryRun, statisticSwitch: make(chan struct{}),
		configmapName: fmt.Sprintf("%s-%s", ConfigmapPrefix, deploy.DeployName),
		configmapData: ranktable, statisticStopped: false, cachedPodNum: 0, taskReplicasTotal: replicasTotal},
		DeployInfo: deploy}
}

func (w *DeployWorker) doWork(pod *apiCoreV1.Pod, podInfo *podIdentifier) (bool, bool) {
	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if pod.CreationTimestamp.Before(&w.DeployCreationTimestamp) {
		// old pod + new worker
		hwlog.RunLog.Infof("syncing '%s' terminated: corresponding job worker is no "+
			"longer exist (basis: job uid + creationTimestamp)", podInfo)
		return true, false
	}

	// check whether pod has used npu
	if used := containerUsedChip(pod); !used {
		hwlog.RunLog.Errorf("pod %s doesn't use npu, so no longer dealing with it", podInfo)
		return true, true
	}
	// scenario check C: if current pod use chip, its' device info may not be ready
	// check basis: limits + annotations
	if (podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate) && !isPodAnnotationsReady(pod,
		podInfo.String()) {
		return false, false
	}
	if w.configmapData.GetStatus() == ConfigmapCompleted {
		hwlog.RunLog.Infof("syncing '%s' terminated: corresponding rank table is completed",
			podInfo)
		return true, true
	}

	// start to sync current pod
	if err := w.syncHandler(pod, podInfo); err != nil {
		hwlog.RunLog.Errorf("error syncing '%s': %s", podInfo, err.Error())
		return true, true
	}
	return true, true
}

// Statistic : no need to add lock here, deviation from true value is acceptable
func (w *DeployWorker) Statistic(stopTime time.Duration) {
	for {
		select {
		case c, ok := <-w.statisticSwitch:
			if !ok {
				hwlog.RunLog.Error(c)
			}
			return
		default:
			if w.taskReplicasTotal == w.cachedPodNum {
				hwlog.RunLog.Infof("rank table build progress for %s/%s is completed",
					w.DeployNamespace, w.DeployName)
				w.CloseStatistic()
				return
			}
			hwlog.RunLog.Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", w.DeployNamespace, w.DeployName, w.taskReplicasTotal, w.cachedPodNum)
			time.Sleep(stopTime)
		}
	}
}
