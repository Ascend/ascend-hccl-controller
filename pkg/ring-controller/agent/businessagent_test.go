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
	_ "hccl-controller/pkg/test-util"
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
