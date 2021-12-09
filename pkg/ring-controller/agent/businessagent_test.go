/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */
package agent

import (
	. "github.com/smartystreets/goconvey/convey"
	_ "hccl-controller/pkg/testtool"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

// TestDeleteWorker test DeleteWorker
func TestDeleteWorker(t *testing.T) {
	Convey("agent DeleteWorker", t, func() {
		bus, _ := NewBusinessAgent(fake.NewSimpleClientset(), nil,
			&Config{PodParallelism: 1}, make(chan struct{}))
		Convey("DeleteWorker businessAgent when exist", func() {
			bus.BusinessWorker["namespace/test"] = new(VCJobWorker)
			DeleteWorker("namespace", "test", bus)
			So(len(bus.BusinessWorker), ShouldEqual, 0)
		})
		Convey("DeleteWorker businessAgent when not exist", func() {
			bus.BusinessWorker["namespace/test1"] = nil
			DeleteWorker("namespace", "test", bus)
			So(len(bus.BusinessWorker), ShouldEqual, 1)
		})
	})
}
