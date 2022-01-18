/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

// Package v1 ranktable version 1
package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	"net"
	"strconv"
	"unicode/utf8"
)

// RankTabler interface to maintain properties
type RankTabler interface {
	// UnmarshalToRankTable Unmarshal json string to RankTable
	UnmarshalToRankTable(jsonString string) error
	// CachePodInfo cache pod info to RankTableV1
	CachePodInfo(pod *apiCoreV1.Pod, instance Instance, rankIndex *int) error
	// RemovePodInfo Remove pod info from RankTable
	RemovePodInfo(namespace string, name string) error
	// SetStatus Set status of RankTableStatus
	SetStatus(status string) error
	// GetStatus Get status of RankTableStatus
	GetStatus() string
	// GetPodNum get pod num
	GetPodNum() int
}

// SetStatus Set status of RankTableStatus
func (r *RankTableStatus) SetStatus(status string) error {
	r.Status = status
	return nil
}

// GetStatus : Get status of RankTableStatus
func (r *RankTableStatus) GetStatus() string {
	return r.Status
}

// UnmarshalToRankTable ： Unmarshal json string to RankTable
func (r *RankTableStatus) UnmarshalToRankTable(jsonString string) error {
	if utf8.RuneCount([]byte(jsonString)) > maximumMemory {
		return fmt.Errorf("out of memory")
	}
	err := json.Unmarshal([]byte(jsonString), &r)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if r.Status != ConfigmapCompleted && r.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}
	return nil
}

// CheckDeviceInfo ：validation of DeviceInfo
func CheckDeviceInfo(instance *Instance) bool {
	if parsedIP := net.ParseIP(instance.ServerID); parsedIP == nil {
		return false
	}
	if len(instance.Devices) == 0 {
		return false
	}
	for _, item := range instance.Devices {

		if value, err := strconv.Atoi(item.DeviceID); err != nil || value < 0 {
			return false
		}
		if parsedIP := net.ParseIP(item.DeviceIP); parsedIP == nil {
			return false
		}
	}
	return true
}

// CachePodInfo : cache pod info to RankTableV1
func (r *RankTable) CachePodInfo(pod *apiCoreV1.Pod, instance Instance, rankIndex *int) error {
	if len(r.GroupList) < 1 {
		return fmt.Errorf("grouplist of ranktable is empty")
	}
	group := r.GroupList[0]
	if err := checkPodCache(group, pod); err != nil {
		return err
	}
	hwlog.RunLog.Infof("instance from pod: %v", instance)
	if !CheckDeviceInfo(&instance) {
		return errors.New("deviceInfo failed the validation")
	}
	group.InstanceList = append(group.InstanceList, &instance)
	*rankIndex++
	return nil
}

// RemovePodInfo : Remove pod info from RankTable
func (r *RankTable) RemovePodInfo(namespace string, podID string) error {
	hasInfoToRemove := false
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
		return fmt.Errorf("no data of pod %s/%s can be removed", namespace, podID)
	}

	return nil
}

func checkPodCache(group *Group, pod *apiCoreV1.Pod) error {
	for _, instance := range group.InstanceList {
		if instance.PodName == pod.Name {
			hwlog.RunLog.Infof("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
			return fmt.Errorf("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
		}
	}
	return nil
}

// GetPodNum get pod num
func (r *RankTable) GetPodNum() int {
	return 0
}
