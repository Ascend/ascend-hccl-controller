/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package v2 ranktable version 2
package v2

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	apiCoreV1 "k8s.io/api/core/v1"

	"hccl-controller/pkg/ring-controller/common"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

// CachePodInfo :Cache pod info to RankTableV2
func (r *RankTable) CachePodInfo(pod *apiCoreV1.Pod, instance v1.Instance, rankIndex *int) error {
	var server Server
	if !v1.CheckDeviceInfo(&instance) {
		return errors.New("deviceInfo failed the validation")
	}
	for _, server := range r.ServerList {
		if server.PodID == instance.PodName {
			return fmt.Errorf("ANOMALY: pod %s/%s is already cached", pod.Namespace, pod.Name)
		}
	}

	// Build new server-level struct from device info
	server.ServerID = instance.ServerID
	server.PodID = instance.PodName
	rankFactor := len(instance.Devices)
	for _, device := range instance.Devices {
		var serverDevice Device
		serverDevice.DeviceID = device.DeviceID
		serverDevice.DeviceIP = device.DeviceIP
		serverDevice.RankID = strconv.Itoa(*rankIndex*rankFactor + len(server.DeviceList))

		server.DeviceList = append(server.DeviceList, &serverDevice)
	}
	if len(server.DeviceList) < 1 {
		return fmt.Errorf("%s/%s get deviceList failed", pod.Namespace, pod.Name)
	}

	r.ServerList = append(r.ServerList, &server)
	sort.Slice(r.ServerList, func(i, j int) bool {
		iRank, err := strconv.ParseInt(r.ServerList[i].DeviceList[0].RankID, common.Decimal, common.BitSize32)
		jRank, err2 := strconv.ParseInt(r.ServerList[j].DeviceList[0].RankID, common.Decimal, common.BitSize32)
		if err != nil || err2 != nil {
			return false
		}
		return iRank < jRank
	})
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

// GetPodNum get pod num
func (r *RankTable) GetPodNum() int {
	serverLen := len(r.ServerList)
	if serverLen == 0 {
		return 0
	}
	return serverLen * len(r.ServerList[0].DeviceList)
}
