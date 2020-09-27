/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
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

// Package controller for controller
package controller

import "time"

const (
	// Key910 to get Configmap
	Key910 = "ring-controller.atlas"
	// Val910 to get Configmap
	Val910 = "ascend-910" // Val910 to get Configmap
	// ResourceName for 910
	ResourceName   = "huawei.com/Ascend910"
	controllerName = "ring-controller"
	// ConfigmapPrefix to get from configmap
	ConfigmapPrefix = "rings-config"
	// ConfigmapCompleted Staus
	ConfigmapCompleted = "completed"
	// ConfigmapInitializing status
	ConfigmapInitializing = "initializing"
	// ConfigmapKey configmap Data Name
	ConfigmapKey = "hccl.json"
	// VolcanoJobNameKey to get job name
	VolcanoJobNameKey = "volcano.sh/job-name"
	// PodJobVersion to get job version
	PodJobVersion = "volcano.sh/job-version"
	// PodDeviceKey Pod annoation Key
	PodDeviceKey = "atlas.kubectl.kubernetes.io/ascend-910-configuration"
	// PodGroupKey to get Group key
	PodGroupKey = "volcano.sh/task-spec"
	// JobRestartPhase restart flage
	JobRestartPhase = "Restarting"
	// EventAdd event add
	EventAdd = "add"
	// EventUpdate event to update
	EventUpdate = "update"
	// EventDelete event to delete
	EventDelete = "delete"
	// BuildStatInterval 1 * time.Minute
	BuildStatInterval = 30 * time.Second
	serverIP          = "serverIp"
	// L1 log level 1
	L1 = 1
	// L2 log level 2
	L2 = 2
	// L3 log level 3
	L3 = 3
	// L4 log level 4
	L4               = 4
	retryMilliSecond = 5
	threeMinutes     = 180
	splitNum         = 4
	decimal          = 10
	two              = 2
	twosecond        = 2 * time.Second
	three            = 3
	four             = 4
	eight            = 8
	status           = 200
)

// RankTable to hccl
type RankTable struct {
	GroupList  []*Group `json:"group_list"`          // hccl group list
	Status     string   `json:"status"`              // get hccl_json status
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
	Devices  []Device `json:"devices"`   // hccl Deviceid
	PodName  string   `json:"pod_name"`  // hccl PodName
	ServerID string   `json:"server_id"` // hccl servceId
}

// Device to hccl
type Device struct {
	DeviceID string `json:"device_id"` // hccl deviceId
	DeviceIP string `json:"device_ip"` // hccl deviceid
}
