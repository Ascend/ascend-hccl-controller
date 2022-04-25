/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */
package agent

import (
	"math"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	apiCorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// TestGetNPUNum test GetNPUNum
func TestGetNPUNum(t *testing.T) {
	convey.Convey("Get NPUNum", t, func() {
		convey.Convey("no npu found", func() {
			c := apiCorev1.Container{Resources: apiCorev1.ResourceRequirements{}}
			val := GetNPUNum(c)
			convey.So(val, convey.ShouldEqual, 0)
		})
		convey.Convey("legal npu number", func() {
			rl := apiCorev1.ResourceList{}
			rl[a910With2CResourceName] = resource.MustParse("1")
			c := apiCorev1.Container{Resources: apiCorev1.ResourceRequirements{Limits: rl}}
			val := GetNPUNum(c)
			convey.So(val, convey.ShouldEqual, 1)
		})
		convey.Convey("illegal npu number, number is too big", func() {
			rl := apiCorev1.ResourceList{}
			tooBigNum := math.MaxInt32 + 1
			rl[a910With2CResourceName] = resource.MustParse(strconv.Itoa(tooBigNum))
			c := apiCorev1.Container{Resources: apiCorev1.ResourceRequirements{Limits: rl}}
			val := GetNPUNum(c)
			convey.So(val, convey.ShouldEqual, math.MaxInt32)
		})
		convey.Convey("illegal npu number, number is too small", func() {
			rl := apiCorev1.ResourceList{}
			tooSmallNum := math.MinInt32 - 1
			rl[a910With2CResourceName] = resource.MustParse(strconv.Itoa(tooSmallNum))
			c := apiCorev1.Container{Resources: apiCorev1.ResourceRequirements{Limits: rl}}
			val := GetNPUNum(c)
			convey.So(val, convey.ShouldEqual, math.MinInt32)
		})
	})
}

// TestPreCheck test preCheck
func TestPreCheck(t *testing.T) {
	convey.Convey("test preCheck", t, func() {
		convey.Convey("obj is not a string", func() {
			obj := struct{}{}
			_, retry := preCheck(obj)
			convey.So(retry, convey.ShouldEqual, true)
		})
		convey.Convey("obj is a empty string", func() {
			obj := ""
			_, retry := preCheck(obj)
			convey.So(retry, convey.ShouldEqual, true)
		})
		convey.Convey("a valid obj", func() {
			obj := "default/test-job/vcjob/add"
			_, retry := preCheck(obj)
			convey.So(retry, convey.ShouldEqual, false)
		})
	})
}

// TestIsReferenceJobSameWithBsnsWorker test isReferenceJobSameWithBsnsWorker
func TestIsReferenceJobSameWithBsnsWorker(t *testing.T) {
	convey.Convey("test isReferenceJobSameWithBsnsWorker", t, func() {
		uuid := "UID-xxxxxxxxxxxxxxx"
		name := "test-name"
		or := []metav1.OwnerReference{
			{UID: types.UID(uuid), Name: name},
		}
		pod := apiCorev1.Pod{}
		pod.OwnerReferences = or
		convey.Convey("the same", func() {
			isSame := isReferenceJobSameWithBsnsWorker(&pod, name, uuid)
			convey.So(isSame, convey.ShouldEqual, true)
		})
		convey.Convey("not same", func() {
			isSame := isReferenceJobSameWithBsnsWorker(&pod, "podName", uuid)
			convey.So(isSame, convey.ShouldEqual, false)
		})
	})
}
