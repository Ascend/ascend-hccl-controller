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

const (
	// ResourceName NPU resource Name
	ResourceName = "huawei.com/Ascend910"
	// ConfigmapCompleted Staus
	ConfigmapCompleted = "completed"
	// ConfigmapInitializing status
	ConfigmapInitializing = "initializing"

	// EventAdd event add
	EventAdd = "add"
	// EventUpdate event to update
	EventUpdate = "update"
	// EventDelete event to delete
	EventDelete = "delete"

	// L1 log level 1
	L1 = 1
	// L2 log level 2
	L2 = 2
	// L3 log level 3
	L3 = 3
	// L4 log level 4
	L4 = 4
)

// RankTableStatus to hccl
type RankTableStatus struct {
	Status string `json:"status"` // get hccl_json status
}

// RankTable to hccl
type RankTable struct {
	RankTableStatus
	GroupList  []*Group `json:"group_list"`          // hccl group list
	GroupCount string   `json:"group_count, string"` // hccl_json grouoCount
}

// Group to hccl
type Group struct {
	InstanceList  []*Instance `json:"instance_list"`          // hccl InstaceList
	GroupName     string      `json:"group_name"`             // hccl GroupName
	DeviceCount   string      `json:"device_count, string"`   // hccl Devicecount
	InstanceCount string      `json:"instance_count, string"` // hccl Instance Count
}

// Instance to hccl
type Instance struct {
	Devices  []Device `json:"devices"`   // hccl Device
	PodName  string   `json:"pod_name"`  // hccl PodName
	ServerID string   `json:"server_id"` // hccl servceId
}

// Device to hccl
type Device struct {
	DeviceID string `json:"device_id"` // hccl deviceId
	DeviceIP string `json:"device_ip"` // hccl deviceIp
}
