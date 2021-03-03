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

// Package v1 ranktable version 1
package v1

import (
	"encoding/json"
	"fmt"

	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// RankTable interface to maintain properties
type RankTabler interface {
	// UnmarshalToRankTable：Unmarshal json string to RankTable
	UnmarshalToRankTable(jsonString string) error
	// CachePodInfo: cache pod info to RankTableV1
	CachePodInfo(pod *apiCoreV1.Pod, deviceInfo string, rankIndex *int) error
	// RemovePodInfo : Remove pod info from RankTable
	RemovePodInfo(namespace string, name string) error
	// SetStatus Set status of RankTableStatus
	SetStatus(status string) error
	// GetStatus: Get status of RankTableStatus
	GetStatus() string
}

// SetStatus Set status of RankTableStatus
func (r *RankTableStatus) SetStatus(status string) error {
	r.Status = status
	return nil
}

// GetStatus: Get status of RankTableStatus
func (r *RankTableStatus) GetStatus() string {
	return r.Status
}

// UnmarshalToRankTable： Unmarshal json string to RankTable
func (r *RankTableStatus) UnmarshalToRankTable(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &r)
	if err != nil {
		return fmt.Errorf("parse configmap data error: %v", err)
	}
	if r.Status != ConfigmapCompleted && r.Status != ConfigmapInitializing {
		return fmt.Errorf("configmap status abnormal: %v", err)
	}
	return nil
}

// CachePodInfo: cache pod info to RankTableV1
func (r *RankTable) CachePodInfo(pod *apiCoreV1.Pod, deviceInfo string, rankIndex *int) error {
	if len(r.GroupList) < 1 {
		return fmt.Errorf("grouplist of ranktable is empty")
	}
	group := r.GroupList[0]
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

func checkPodCache(group *Group, pod *apiCoreV1.Pod) bool {
	for _, instance := range group.InstanceList {
		if instance.PodName == pod.Name {
			klog.V(L3).Infof("ANOMALY: pod %s/%s is already cached", pod.Namespace,
				pod.Name)
			return true
		}
	}
	return false
}
