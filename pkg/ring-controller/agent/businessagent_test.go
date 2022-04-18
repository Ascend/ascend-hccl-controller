/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */
package agent

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes/fake"

	_ "hccl-controller/pkg/testtool"
)

// TestDeleteWorker test DeleteWorker
func TestDeleteWorker(t *testing.T) {
	convey.Convey("agent DeleteWorker", t, func() {
		bus, _ := NewBusinessAgent(fake.NewSimpleClientset(), nil,
			&Config{PodParallelism: 1}, make(chan struct{}))
		convey.Convey("DeleteWorker businessAgent when exist", func() {
			bus.BusinessWorker["namespace/test"] = new(VCJobWorker)
			DeleteWorker("namespace", "test", bus)
			convey.So(len(bus.BusinessWorker), convey.ShouldEqual, 0)
		})
		convey.Convey("DeleteWorker businessAgent when not exist", func() {
			bus.BusinessWorker["namespace/test1"] = nil
			DeleteWorker("namespace", "test", bus)
			convey.So(len(bus.BusinessWorker), convey.ShouldEqual, 1)
		})
	})
}
