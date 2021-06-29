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

// Package v2 ranktable version 2
package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	apiCoreV1 "k8s.io/api/core/v1"
	"strconv"
)

// CachePodInfo :Cache pod info to RankTableV2
func (r *RankTable) CachePodInfo(pod *apiCoreV1.Pod, deviceInfo string, rankIndex *int) error {
	var instance v1.Instance
	var server Server

	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		return fmt.Errorf("parse annotation of pod %s/%s error: %v", pod.Namespace, pod.Name, err)
	}
	if !v1.CheckDeviceInfo(&instance) {
		return errors.New("deviceInfo failed the validation")
	}
	for _, server := range r.ServerList {
		if server.PodID == instance.PodName {
			return fmt.Errorf("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
		}
	}

	rankFactor := len(instance.Devices)

	// Build new server-level struct from device info
	server.ServerID = instance.ServerID
	server.PodID = instance.PodName

	for _, device := range instance.Devices {
		var serverDevice Device
		serverDevice.DeviceID = device.DeviceID
		serverDevice.DeviceIP = device.DeviceIP
		serverDevice.RankID = strconv.Itoa(*rankIndex*rankFactor + len(server.DeviceList))

		server.DeviceList = append(server.DeviceList, &serverDevice)
	}

	r.ServerList = append(r.ServerList, &server)
	r.ServerCount = strconv.Itoa(len(r.ServerList))
	*rankIndex++
	return nil
}

// RemovePodInfo :Remove pod info from RankTableV2
func (r *RankTable) RemovePodInfo(namespace string, podID string) error {
	hasInfoToRemove := false
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
		return fmt.Errorf("no data of pod %s/%s can be removed", namespace, podID)
	}
	r.ServerList = serverList
	r.ServerCount = strconv.Itoa(len(r.ServerList))

	return nil
}
