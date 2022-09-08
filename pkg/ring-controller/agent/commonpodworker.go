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

package agent

import (
	"fmt"
	"time"

	apiCoreV1 "k8s.io/api/core/v1"

	"hccl-controller/pkg/hwlog"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// NewCommonPodWorker ： to create Pod Worker
func NewCommonPodWorker(agent *BusinessAgent, job CommonPodInfo, ranktable v1.RankTabler,
	replicasTotal int32, labelKey, labelVal string) *CommonPodWorker {
	return &CommonPodWorker{
		CommonPodWorkerInfo: CommonPodWorkerInfo{
			kubeclientset:     agent.KubeClientSet,
			podsIndexer:       agent.PodsIndexer,
			recorder:          agent.recorder,
			dryRun:            agent.dryRun,
			statisticSwitch:   make(chan struct{}),
			configmapName:     fmt.Sprintf("%s-%s", ConfigmapPrefix, job.Name),
			configmapData:     ranktable,
			statisticStopped:  false,
			cachedPodNum:      0,
			taskReplicasTotal: replicasTotal,
			labelKey:          labelKey,
			labelVal:          labelVal},
		CommonPodInfo: job,
	}
}

func (w *CommonPodWorker) doWork(pod *apiCoreV1.Pod, podInfo *podIdentifier) (forgetQueue, retry bool) {
	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if pod.CreationTimestamp.Before(&w.CreationTimestamp) {
		// old pod + new worker
		hwlog.Infof("syncing '%s' terminated: corresponding job worker is no "+
			"longer exist (basis: job uid + creationTimestamp)", podInfo)
		return true, false
	}
	// scenario check C: if current pod use chip, its' device info may not be ready
	// check basis: limits + annotations
	if (podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate) && !isPodAnnotationsReady(pod,
		podInfo.String()) {
		return false, false
	}
	if configmapComplete :=
		w.configmapData.GetStatus() == ConfigmapCompleted; configmapComplete {
		hwlog.Infof("syncing '%s' terminated: corresponding rank table is completed",
			podInfo)
		return true, true
	}

	// start to sync current pod
	if err := w.syncHandler(pod, podInfo); err != nil {
		hwlog.Errorf("error syncing '%s': %s", podInfo, err.Error())
		return true, true
	}
	return true, true
}

// Statistic : no need to add lock here, deviation from true value is acceptable
func (w *CommonPodWorker) Statistic(stopTime time.Duration) {
	for {
		select {
		case c, ok := <-w.statisticSwitch:
			if !ok {
				hwlog.Error(c)
			}
			return
		default:
			if w.taskReplicasTotal == w.cachedPodNum {
				hwlog.Infof("rank table build progress for %s/%s is completed",
					w.Namespace, w.Name)
				w.CloseStatistic()
				return
			}
			hwlog.Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", w.Namespace, w.Name, w.taskReplicasTotal, w.cachedPodNum)
			time.Sleep(stopTime)
		}
	}
}
