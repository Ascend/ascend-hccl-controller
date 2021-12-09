/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package v2

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	_ "hccl-controller/pkg/testtool"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	"testing"
)

// TestCachePodInfo test  CachePodInfo
func TestCachePodInfo(t *testing.T) {
	Convey("TestRankTableV2 CachePodInfo", t, func() {
		var serverList []*Server
		fake := &RankTable{ServerCount: strconv.Itoa(len(serverList)), ServerList: serverList,
			RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
		po := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		rank := 1
		const (
			podString = "{\"pod_name\":\"test1\",\"server_id\":\"0.0.0.0\"," +
				"\"devices\":[{\"device_id\":\"0\",\"device_ip\":\"0.0.0.0\"}]}"
			RankNumExpect = 2
		)

		Convey("CachePodInfo() should return err == nil when Normal ", func() {
			err := fake.CachePodInfo(po, podString, &rank)
			So(err, ShouldEqual, nil)
			So(rank, ShouldEqual, RankNumExpect)
			deviceIP := fake.ServerList[0].DeviceList[0].DeviceIP
			So(deviceIP, ShouldEqual, "0.0.0.0")
		})

		Convey("CachePodInfo() should return err != nil when deviceInfo is wrong", func() {
			err := fake.CachePodInfo(po, "{\"pod_name\":\"test1\",\"server_id\":}", &rank)
			So(err, ShouldNotEqual, nil)
			So(rank, ShouldEqual, 1)
		})

		Convey("CachePodInfo() should return err != nil when cached ", func() {
			fake.CachePodInfo(po, podString, &rank)
			po2 := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
			err := fake.CachePodInfo(po2, podString, &rank)
			So(err, ShouldNotEqual, nil)
			So(rank, ShouldEqual, RankNumExpect)
		})
	})
}

// TestRemovePodInfo test RemovePodInfo
func TestRemovePodInfo(t *testing.T) {
	Convey("TestRankTableV2 RemovePodInfo", t, func() {
		var serverList []*Server
		fake := &RankTable{ServerCount: strconv.Itoa(len(serverList)), ServerList: serverList,
			RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
		po := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		rank := 1
		const podString = "{\"pod_name\":\"test1\",\"server_id\":\"0.0.0.0\"," +
			"\"devices\":[{\"device_id\":\"0\",\"device_ip\":\"0.0.0.1\"}]}"
		Convey("RemovePodInfo() should return err == nil when Normal", func() {
			fake.CachePodInfo(po, podString, &rank)
			err := fake.RemovePodInfo("", "test1")
			So(err, ShouldEqual, nil)
			So(len(fake.ServerList), ShouldEqual, 0)

		})

		Convey("RemovePodInfo() should return err != nil when podName !contain GroupList ", func() {
			fake.CachePodInfo(po, podString, &rank)
			err := fake.RemovePodInfo("", "1")
			So(err, ShouldNotEqual, nil)
			So(len(fake.ServerList), ShouldEqual, 1)
		})
	})
}
