/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package v2 for v2 rank table test
package v2

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	_ "hccl-controller/pkg/testtool"
)

// TestCachePodInfo test  CachePodInfo
func TestCachePodInfo(t *testing.T) {
	fmt.Println("TestRankTableV2 CachePodInfo")
	po := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
	rank := 1
	const (
		podString = `{"pod_name":"test1","server_id":"0.0.0.0","devices":[{"device_id":"0",
"device_ip":"0.0.0.0"}]}`
		RankNumExpect = 2
	)
	var instance v1.Instance
	if err := json.Unmarshal([]byte(podString), &instance); err != nil {
		instance = v1.Instance{}
	}

	fmt.Println("CachePodInfo() should return err == nil when Normal ")
	fake := &RankTable{ServerCount: "0", ServerList: []*Server{},
		RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
	err := fake.CachePodInfo(po, instance, &rank)
	assert.Equal(t, nil, err)
	assert.Equal(t, RankNumExpect, rank)
	deviceIP := fake.ServerList[0].DeviceList[0].DeviceIP
	assert.Equal(t, "0.0.0.0", deviceIP)

	fmt.Println("CachePodInfo() should return err != nil when cached ")
	fake = &RankTable{ServerCount: "0", ServerList: []*Server{},
		RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
	rank = 1
	fake.CachePodInfo(po, instance, &rank)
	po2 := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
	err = fake.CachePodInfo(po2, instance, &rank)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, RankNumExpect, rank)

	fmt.Println("CachePodInfo() should return err != nil when deviceInfo is wrong")
	rank = 1
	if err = json.Unmarshal([]byte(`{"pod_name":"test1","server_id":}`), &instance); err != nil {
		instance = v1.Instance{}
	}
	fake = &RankTable{ServerCount: "0", ServerList: []*Server{},
		RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
	err = fake.CachePodInfo(po, instance, &rank)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, rank)
}

// TestRemovePodInfo test RemovePodInfo
func TestRemovePodInfo(t *testing.T) {
	convey.Convey("TestRankTableV2 RemovePodInfo", t, func() {
		var serverList []*Server
		fake := &RankTable{ServerCount: strconv.Itoa(len(serverList)), ServerList: serverList,
			RankTableStatus: v1.RankTableStatus{Status: v1.ConfigmapInitializing}, Version: "1.0"}
		po := &apiCoreV1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		rank := 1
		const podString = `{"pod_name":"test1","server_id":"0.0.0.0","devices":[{"device_id":"0",
"device_ip":"0.0.0.1"}]}`
		var instance v1.Instance
		if err := json.Unmarshal([]byte(podString), &instance); err != nil {
			instance = v1.Instance{}
		}
		convey.Convey("RemovePodInfo() should return err == nil when Normal", func() {
			fake.CachePodInfo(po, instance, &rank)
			err := fake.RemovePodInfo("", "test1")
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(len(fake.ServerList), convey.ShouldEqual, 0)

		})

		convey.Convey("RemovePodInfo() should return err != nil when podName !contain GroupList ", func() {
			fake.CachePodInfo(po, instance, &rank)
			err := fake.RemovePodInfo("", "1")
			convey.So(err, convey.ShouldNotEqual, nil)
			convey.So(len(fake.ServerList), convey.ShouldEqual, 1)
		})
	})
}
