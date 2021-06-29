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

package v1

import (
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestUnmarshalToRankTable test UnmarshalToRankTable
func TestUnmarshalToRankTable(t *testing.T) {
	Convey("TestRankTableV1 UnmarshalToRankTable", t, func() {
		r := &RankTableStatus{}
		Convey("UnmarshalToRankTable() should return err == nil &&"+
			" r.status == ConfigmapInitializing when Normal", func() {
			err := r.UnmarshalToRankTable("{\"status\":\"initializing\"}")
			So(err, ShouldEqual, nil)
			So(r.Status, ShouldEqual, ConfigmapInitializing)
		})
		Convey("UnmarshalToRankTable should return err != nil when "+
			"jobString == \"status\":\"initializing\" ", func() {
			err := r.UnmarshalToRankTable("\"status\":\"initializing\"")
			So(err, ShouldNotEqual, nil)
			So(r.Status, ShouldEqual, "")
		})
		Convey("UnmarshalToRankTable should return err != nil when jobString == "+
			"{\"status\":\"xxxxx\"} ", func() {
			err := r.UnmarshalToRankTable("{\"status\":\"xxxxx\"}")
			So(err, ShouldNotEqual, nil)
		})
	})

}

//
func TestCheckDeviceInfo(t *testing.T) {
	Convey("TestRankTableV1 TestCheckDeviceInfo", t, func() {
		instance := Instance{
			Devices:  []Device{{DeviceID: "2", DeviceIP: "51.38.67.98"}, {DeviceID: "3", DeviceIP: "51.38.67.93"}},
			PodName:  "podname",
			ServerID: "51.38.67.98",
		}
		Convey("CheckDeviceInfo() should return true when Normal", func() {
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, true)
		})
		Convey("CheckDeviceInfo() should return false when ServerID  is not an IP address", func() {
			instance.ServerID = "51.38.67.98s"
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
			instance.Devices[0].DeviceIP = "51w.38.67.98s"
			isOk := CheckDeviceInfo(&instance)
			So(isOk, ShouldEqual, false)
		})

	})
}

// TestCachePodInfo test CachePodInfo
func TestCachePodInfo(t *testing.T) {
	Convey("TestRankTableV1 TestCachePodInfo", t, func() {
		group := &Group{GroupName: "t1", DeviceCount: "1", InstanceCount: "1", InstanceList: []*Instance(nil)}
		groupList := append([]*Group(nil), group)
		fake := &RankTable{RankTableStatus: RankTableStatus{Status: ConfigmapInitializing}, GroupCount: "1",
			GroupList: groupList}
		po := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		rank := 1
		const (
			podString = "{\"pod_name\":\"0\",\"server_id\":\"0.0.0.0\"," +
				"\"devices\":[{\"device_id\":\"0\",\"device_ip\":\"0.0.0.0\"}]}"
			RankNumExpect = 2
		)

		Convey("CachePodInfo() should return err == nil when Normal ", func() {
			err := fake.CachePodInfo(po, podString, &rank)
			So(err, ShouldEqual, nil)
			So(rank, ShouldEqual, RankNumExpect)
			deviceIP := fake.GroupList[0].InstanceList[0].Devices[0].DeviceIP
			So(deviceIP, ShouldEqual, "0.0.0.0")
		})

		Convey("CachePodInfo() should return err != nil when podName == group.Instance.PodName", func() {
			fake.CachePodInfo(po, podString, &rank)
			po2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "0"}}
			err := fake.CachePodInfo(po2, podString, &rank)
			So(err, ShouldNotEqual, nil)
			So(rank, ShouldEqual, RankNumExpect)
		})

		Convey("CachePodInfo() should return err != nil when deviceInfo is wrong", func() {
			err := fake.CachePodInfo(po, "{\"pod_name\":\"0\",\"server_id\":}", &rank)
			So(err, ShouldNotEqual, nil)
			So(rank, ShouldEqual, 1)
		})

		Convey("CachePodInfo() should retrun err != nil when len(GroupCount) <1 ", func() {
			fake := &RankTable{RankTableStatus: RankTableStatus{Status: ConfigmapInitializing},
				GroupCount: "1", GroupList: nil}
			err := fake.CachePodInfo(nil, "", nil)
			So(err, ShouldNotEqual, nil)
		})

	})
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
		const podString = "{\"pod_name\":\"test1\",\"server_id\":\"0.0.0.0\"," +
			"\"devices\":[{\"device_id\":\"0\",\"device_ip\":\"127.0.0.1\"}]}"
		Convey("RemovePodInfo() should return err == nil when Normal", func() {
			fake.CachePodInfo(po, podString, &rank)
			err := fake.RemovePodInfo("", po.Name)
			So(err, ShouldEqual, nil)
			So(len(fake.GroupList[0].InstanceList), ShouldEqual, 0)
		})

		Convey("RemovePodInfo() should return err != nil when podName !contain GroupList ", func() {
			fake.CachePodInfo(po, podString, &rank)
			err := fake.RemovePodInfo("", "0")
			So(err, ShouldNotEqual, nil)
			So(len(fake.GroupList[0].InstanceList), ShouldEqual, 1)
		})

	})
}
