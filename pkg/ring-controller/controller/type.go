/*
 * Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
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

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
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
	oneMinitue       = 60
)

var (
	// Hccl.json template version
	JsonVersion = "v2"
)

// RankTable interface to maintain properties
type RankTable interface {
	unmarshalToRankTable(jsonString string) error
}

// RankTableV1 to hccl
type RankTableV1 struct {
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

// RankTableV2 to hccl
type RankTableV2 struct {
	ServerList  []*Server `json:"server_list"`  // hccl_json server list
	ServerCount string    `json:"server_count"` // hccl_json server count
	Status      string    `json:"status"`       // hccl_json status
	Version     string    `json:"version"`      // hccl_json version
}

// Server to hccl
type Server struct {
	DeviceList []*DeviceV2 `json:"device"`    // device list in each server
	ServerID   string      `json:"server_id"` // server id, represented by ip address
	PodID      string      `json:"-"`         // pod id, equal to the last integer of pod name
}

// DeviceV2 to hccl
type DeviceV2 struct {
	DeviceID string `json:"device_id"` // device id
	DeviceIP string `json:"device_ip"` // device ip
	RankID   string `json:"rank_id"`   // rank id
}

// Unmarshal json string to RankTableV1
func (configmapDataV1 *RankTableV1) unmarshalToRankTable(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &configmapDataV1)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if configmapDataV1.Status != ConfigmapCompleted && configmapDataV1.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}
	return nil
}

// Unmarshal json string to RankTableV2
func (configmapDataV2 *RankTableV2) unmarshalToRankTable(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &configmapDataV2)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if configmapDataV2.Status != ConfigmapCompleted && configmapDataV2.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}
	return nil
}
