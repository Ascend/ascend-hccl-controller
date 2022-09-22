/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package v1

const (
	// ResourceName NPU resource Name
	ResourceName = "huawei.com/Ascend910"
	// ConfigmapCompleted Status
	ConfigmapCompleted = "completed"
	// ConfigmapInitializing status
	ConfigmapInitializing = "initializing"

	// configmap max data size 1MB
	cmDataMaxMemory = 1024 * 1024
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
