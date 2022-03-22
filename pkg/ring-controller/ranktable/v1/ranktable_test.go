/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package v1
package v1

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "hccl-controller/pkg/testtool"
)

// TestUnmarshalToRankTable test UnmarshalToRankTable
func TestUnmarshalToRankTable(t *testing.T) {
	Convey("TestRankTableV1 UnmarshalToRankTable", t, func() {
		r := &RankTableStatus{}
		Convey("UnmarshalToRankTable() should return err == nil &&"+
			" r.status == ConfigmapInitializing when Normal", func() {
			err := r.UnmarshalToRankTable(`{"status":"initializing"}`)
			So(err, ShouldEqual, nil)
			So(r.Status, ShouldEqual, ConfigmapInitializing)
		})
		Convey("UnmarshalToRankTable should return err != nil when "+
			"jobString == "+`"status": "initializing"`, func() {
			err := r.UnmarshalToRankTable(`"status": "initializing"`)
			So(err, ShouldNotEqual, nil)
			So(r.Status, ShouldEqual, "")
		})
		Convey("UnmarshalToRankTable should return err != nil when jobString == "+
			`{"status":"xxxxx"} `, func() {
			err := r.UnmarshalToRankTable(`{"status":"xxxxx"}`)
			So(err, ShouldNotEqual, nil)
		})
	})

}

//
func TestCheckDeviceInfo(t *testing.T) {
	Convey("TestRankTableV1 TestCheckDeviceInfo", t, func() {
		instance := Instance{
			Devices:  []Device{{DeviceID: "2", DeviceIP: "0.0.0.0"}, {DeviceID: "3", DeviceIP: "0.0.0.0"}},
			PodName:  "podname",
			ServerID: "0.0.0.0",
		}
		Convey("CheckDeviceInfo() should return true when Normal", func() {
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, true)
		})
		Convey("CheckDeviceInfo() should return false when ServerID  is not an IP address", func() {
			instance.ServerID = "127.0.0.1s"
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, false)
		})
		Convey("CheckDeviceInfo() should return false when DeviceID  is less than zero", func() {
			instance.Devices[0].DeviceIP = "-1"
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, false)
		})
		Convey("CheckDeviceInfo() should return false when Devices  is empty", func() {
			instance.Devices = []Device{}
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, false)
		})
		Convey("CheckDeviceInfo() should return false when DeviceIP  is not an IP address", func() {
			instance.Devices[0].DeviceIP = "127w.0.0.1s"
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, false)
		})

	})
}

// TestCachePodInfo test CachePodInfo
func TestCachePodInfo(t *testing.T) {
	fmt.Println("TestRankTableV1 TestCachePodInfo")
	group := &Group{GroupName: "t1", DeviceCount: "1", InstanceCount: "1", InstanceList: []*Instance(nil)}
	groupList := append([]*Group(nil), group)
	fake := &RankTable{RankTableStatus: RankTableStatus{Status: ConfigmapInitializing}, GroupCount: "1",
		GroupList: groupList}
	po := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
	rank := 1
	const (
		podString     = `{"pod_name":"0","server_id":"0.0.0.0","devices":[{"device_id":"0","device_ip":"0.0.0.0"}]}`
		RankNumExpect = 2
	)
	var instance Instance
	if err := json.Unmarshal([]byte(podString), &instance); err != nil {
		instance = Instance{}
	}

	fmt.Println("CachePodInfo() should return err == nil when Normal ")
	err := fake.CachePodInfo(po, instance, &rank)
	assert.Equal(t, nil, err)
	assert.Equal(t, RankNumExpect, rank)
	deviceIP := fake.GroupList[0].InstanceList[0].Devices[0].DeviceIP
	assert.Equal(t, "0.0.0.0", deviceIP)

	fmt.Println("CachePodInfo() should return err != nil when podName == group.Instance.PodName")
	rank = 1
	fake.CachePodInfo(po, instance, &rank)
	po2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "0"}}
	err = fake.CachePodInfo(po2, instance, &rank)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, RankNumExpect, rank)

	fmt.Println("CachePodInfo() should return err != nil when deviceInfo is wrong")
	rank = 1
	if err = json.Unmarshal([]byte(`{"pod_name":"0","server_id":}`), &instance); err != nil {
		instance = Instance{}
	}
	err = fake.CachePodInfo(po, instance, &rank)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, rank)

	fmt.Println("CachePodInfo() should return err != nil when len(GroupCount) <1 ")
	fake = &RankTable{RankTableStatus: RankTableStatus{Status: ConfigmapInitializing},
		GroupCount: "1", GroupList: nil}
	if err = json.Unmarshal([]byte(""), &instance); err != nil {
		instance = Instance{}
	}
	err = fake.CachePodInfo(nil, instance, nil)
	assert.NotEqual(t, nil, err)
}

// TestRemovePodInfo test RemovePodInfo
func TestRemovePodInfo(t *testing.T) {

	Convey("TestRankTableV1 RemovePodInfo", t, func() {
		group := &Group{GroupName: "t1", DeviceCount: "1", InstanceCount: "1", InstanceList: []*Instance(nil)}
		groupList := append([]*Group(nil), group)
		fake := &RankTable{RankTableStatus: RankTableStatus{Status: ConfigmapInitializing},
			GroupCount: "1", GroupList: groupList}
		po := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		rank := 1
		const podString = `{"pod_name":"test1","server_id":"0.0.0.0","devices":[{"device_id":"0",
"device_ip":"127.0.0.1"}]}`
		var instance Instance
		if err := json.Unmarshal([]byte(podString), &instance); err != nil {
			instance = Instance{}
		}
		Convey("RemovePodInfo() should return err == nil when Normal", func() {
			fake.CachePodInfo(po, instance, &rank)
			err := fake.RemovePodInfo("", po.Name)
			So(err, ShouldEqual, nil)
			So(len(fake.GroupList[0].InstanceList), ShouldEqual, 0)
		})

		Convey("RemovePodInfo() should return err != nil when podName !contain GroupList ", func() {
			fake.CachePodInfo(po, instance, &rank)
			err := fake.RemovePodInfo("", "0")
			So(err, ShouldNotEqual, nil)
			So(len(fake.GroupList[0].InstanceList), ShouldEqual, 1)
		})

	})
}
