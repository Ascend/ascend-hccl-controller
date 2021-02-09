/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
	"k8s.io/klog"
	"strconv"
	"strings"

	apiCoreV1 "k8s.io/api/core/v1"
)

// RankTable interface to maintain properties
type RankTable interface {
	unmarshalToRankTable(jsonString string) error
	cachePodInfo(pod *apiCoreV1.Pod, deviceInfo string) error
	removePodInfo(namespace string, name string) error
	setStatus(status string) error
	getStatus() string
}

// RankTableStatus to hccl
type RankTableStatus struct {
	Status string `json:"status"` // get hccl_json status
}

// RankTableV1 to hccl
type RankTableV1 struct {
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
	Devices  []Device `json:"devices"`   // hccl devices
	PodName  string   `json:"pod_name"`  // hccl PodName
	ServerID string   `json:"server_id"` // hccl servceId
}

// Device to hccl
type Device struct {
	DeviceID string `json:"device_id"` // hccl deviceId
	DeviceIP string `json:"device_ip"` // hccl deviceIP
}

// RankTableV2 to hccl
type RankTableV2 struct {
	RankTableStatus
	ServerList  []*Server `json:"server_list"`  // hccl_json server list
	ServerCount string    `json:"server_count"` // hccl_json server count
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
	Device
	RankID string `json:"rank_id"` // rank id
}

// Unmarshal json string to RankTable
func (r *RankTableStatus) unmarshalToRankTable(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &r)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if r.Status != ConfigmapCompleted && r.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}
	return nil
}

// Cache pod info to RankTableV1
func (r *RankTableV1) cachePodInfo(pod *apiCoreV1.Pod, deviceInfo string) error {
	if len(r.GroupList) < 1 {
		return fmt.Errorf("grouplist of ranktable is empty")
	}
	group := r.GroupList[0]
	if group.GroupName != pod.Annotations[PodGroupKey] {
		return nil
	}
	done := checkPodCache(group, pod)
	if done {
		return nil
	}
	var instance Instance

	// if pod use D chip, cache its info
	klog.V(L3).Infof("devicedInfo  from pod => %v", deviceInfo)
	err := json.Unmarshal([]byte(deviceInfo), &instance)
	klog.V(L3).Infof("instace  from pod => %v", instance)
	if err != nil {
		return fmt.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}

	group.InstanceList = append(group.InstanceList, &instance)

	return nil
}

// Cache pod info to RankTableV2
func (r *RankTableV2) cachePodInfo(pod *apiCoreV1.Pod, deviceInfo string) error {
	var instance Instance
	var server Server

	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		return fmt.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}
	rankFactor := len(instance.Devices)

	// Build new server-level struct from device info
	server.ServerID = instance.ServerID
	server.PodID = instance.PodName
	podID, err := strconv.Atoi(server.PodID)
	if err != nil {
		return fmt.Errorf("parse name of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}

	for _, device := range instance.Devices {
		var serverDevice DeviceV2
		serverDevice.DeviceID = device.DeviceID
		serverDevice.DeviceIP = device.DeviceIP
		serverDevice.RankID = strconv.Itoa(podID*rankFactor + len(server.DeviceList))

		server.DeviceList = append(server.DeviceList, &serverDevice)
	}

	r.ServerList = append(r.ServerList, &server)
	r.ServerCount = strconv.Itoa(len(r.ServerList))

	return nil
}

// Remove pod info from RankTableV1
func (r *RankTableV1) removePodInfo(namespace string, name string) error {
	hasInfoToRemove := false

	// Get last bit of pod name as podID
	splited := strings.Split(name, "-")
	podID := splited[len(splited)-1]
	for _, group := range r.GroupList {
		for idx, instance := range group.InstanceList {
			// current pod's info is already cached, start to remove
			if instance.PodName == podID {
				length := len(group.InstanceList)
				group.InstanceList[idx] = group.InstanceList[length-1]
				group.InstanceList = group.InstanceList[:length-1]
				hasInfoToRemove = true
				break
			}
		}
		if hasInfoToRemove {
			break
		}
	}
	if !hasInfoToRemove {
		klog.V(L3).Infof("no data of pod %s/%s can be removed", namespace, name)
		return nil
	}

	return nil
}

// Remove pod info from RankTableV2
func (r *RankTableV2) removePodInfo(namespace string, name string) error {
	hasInfoToRemove := false

	// Get last bit of pod name as podID
	splited := strings.Split(name, "-")
	podID := splited[len(splited)-1]
	serverList := r.ServerList
	for idx, server := range serverList {
		if server.PodID == podID {
			length := len(serverList)
			serverList[idx] = serverList[length-1]
			serverList = serverList[:length-1]
			hasInfoToRemove = true
			break
		}
	}

	if !hasInfoToRemove {
		klog.V(L3).Infof("no data of pod %s/%s can be removed", namespace, name)
		return nil
	}
	r.ServerCount = strconv.Itoa(len(r.ServerList))

	return nil
}

// Set status of RankTableStatus
func (r *RankTableStatus) setStatus(status string) error {
	r.Status = status
	return nil
}

// Get status of RankTableStatus
func (r *RankTableStatus) getStatus() string {
	return r.Status
}
