/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
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
