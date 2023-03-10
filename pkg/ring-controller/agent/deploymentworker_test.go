/*
 * Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

// TestDeployWorkerStatistic test DeployWorker_Statistic
func TestDeployWorkerStatistic(t *testing.T) {
	Convey("agent VCJobWorker_Statistic", t, func() {
		d := &DeployWorker{WorkerInfo: WorkerInfo{statisticSwitch: make(chan struct{}), statisticStopped: false}}
		const (
			TaskRep   = 2
			SleepTime = 3
		)

		Convey(" chan will return when chan close ", func() {
			d.taskReplicasTotal = TaskRep
			d.cachedPodNum = 1
			go func() {
				time.Sleep(SleepTime * time.Second)
				d.CloseStatistic()
			}()
			d.Statistic(1 * time.Second)
		})

		Convey(" chan will return when taskReplicasTotal==cachedPodNum ", func() {
			const CachePod = 2
			d.taskReplicasTotal = TaskRep
			d.cachedPodNum = CachePod
			d.Statistic(1 * time.Second)
		})
	})
}
