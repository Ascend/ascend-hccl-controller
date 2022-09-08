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

package model

import (
	"context"
	"fmt"
	"time"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	v1common "hccl-controller/pkg/apis/training/common"
	v1medal "hccl-controller/pkg/apis/training/medal/v1"
	v1mpi "hccl-controller/pkg/apis/training/mpi/v1"
	v1tf "hccl-controller/pkg/apis/training/tensorflow/v1"
	agent2 "hccl-controller/pkg/ring-controller/agent"
	"hccl-controller/pkg/ring-controller/common"
	v1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	v2 "hccl-controller/pkg/ring-controller/ranktable/v2"
)

// ResourceEventHandler to define same func, controller to use this function to finish something.
type ResourceEventHandler interface {
	EventAdd(tagentInterface *agent2.BusinessAgent) error
	EventUpdate(tagentInterface *agent2.BusinessAgent) error
	GenerateGrouplist() ([]*v1.Group, int32, error)
	GetReplicas() string
	GetCacheIndex() cache.Indexer
	GetModelKey() string
}

// GetModelKey : return model key.
func (model *commonWorkloadInfo) GetModelKey() string {
	return model.key
}

// GetCacheIndex : return CacheIndex
func (model *commonWorkloadInfo) GetCacheIndex() cache.Indexer {
	return model.cacheIndexer
}

// checkCMCreation check configmap
func checkCMCreation(kubeClientSet kubernetes.Interface,
	config *agent2.Config, namespace, name, labelKey, labelVal string) (
	*apiCoreV1.ConfigMap, error) {
	var cm *apiCoreV1.ConfigMap
	err := wait.PollImmediate(time.Duration(config.CmCheckTimeout)*time.Second,
		time.Duration(config.CmCheckTimeout)*time.Second,
		func() (bool, error) {
			var errTmp error

			cm, errTmp = kubeClientSet.CoreV1().ConfigMaps(namespace).
				Get(context.TODO(), fmt.Sprintf("%s-%s", agent2.ConfigmapPrefix, name), metav1.GetOptions{})
			if errTmp != nil {
				if errors.IsNotFound(errTmp) {
					return false, nil
				}
				return true, fmt.Errorf("get configmap error: %v", errTmp)
			}
			return true, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap for job %s/%s: %v", namespace, name, err)
	}
	label910, exist := (*cm).Labels[labelKey]
	if !exist || (exist && label910 != labelVal) {
		return nil, fmt.Errorf("invalid configmap label" + label910)
	}

	return cm, nil
}

// Factory : to generate model
func Factory(obj interface{}, indexers map[string]cache.Indexer,
	eventType, labelKey, labelVal string) (ResourceEventHandler, error) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("object has no meta: %v", err)
	}
	key := metaData.GetName() + "/" + eventType
	if len(metaData.GetNamespace()) > 0 {
		key = metaData.GetNamespace() + "/" + metaData.GetName() + "/" + eventType
	}
	var model ResourceEventHandler

	typeList := []string{common.DeploymentType, common.K8sJobType, common.MedalType, common.MedalType, common.TfType}
	for _, typeName := range typeList {
		if _, ok := indexers[typeName]; !ok {
			return nil, fmt.Errorf("The key: %s, does not exist err %v ", typeName, ok)
		}
	}

	switch t := obj.(type) {
	case *appsV1.Deployment:
		model = &CommonWorkload{
			commonWorkloadInfo: commonWorkloadInfo{
				key:          key,
				cacheIndexer: indexers[common.DeploymentType],
				labelKey:     labelKey,
				labelVal:     labelVal},
			containers: t.Spec.Template.Spec.Containers,
			replicas:   *t.Spec.Replicas,
			CommonPodInfo: agent2.CommonPodInfo{
				Namespace:         t.Namespace,
				Name:              t.Name,
				CreationTimestamp: t.CreationTimestamp}}
	case *batchV1.Job:
		model = &CommonWorkload{
			commonWorkloadInfo: commonWorkloadInfo{key: key,
				cacheIndexer: indexers[common.K8sJobType],
				labelKey:     labelKey,
				labelVal:     labelVal},
			containers: t.Spec.Template.Spec.Containers,
			replicas:   *t.Spec.Parallelism,
			CommonPodInfo: agent2.CommonPodInfo{
				Namespace:         t.Namespace,
				Name:              t.Name,
				CreationTimestamp: t.CreationTimestamp}}
	case *v1medal.MedalJob:
		model = &CommonWorkload{
			commonWorkloadInfo: commonWorkloadInfo{
				key:          key,
				cacheIndexer: indexers[common.MedalType],
				labelKey:     labelKey,
				labelVal:     labelVal},
			containers: t.Spec.MedalReplicaSpecs[v1common.ReplicaTypeWorker].Template.Spec.Containers,
			replicas:   *t.Spec.MedalReplicaSpecs[v1common.ReplicaTypeWorker].Replicas,
			CommonPodInfo: agent2.CommonPodInfo{
				Namespace:         t.Namespace,
				Name:              t.Name,
				CreationTimestamp: t.CreationTimestamp}}
	case *v1mpi.MPIJob:
		model = &CommonWorkload{
			commonWorkloadInfo: commonWorkloadInfo{
				key:          key,
				cacheIndexer: indexers[common.MedalType],
				labelKey:     labelKey,
				labelVal:     labelVal},
			containers: t.Spec.MPIReplicaSpecs[v1common.ReplicaTypeWorker].Template.Spec.Containers,
			replicas:   *t.Spec.MPIReplicaSpecs[v1common.ReplicaTypeWorker].Replicas,
			CommonPodInfo: agent2.CommonPodInfo{
				Namespace:         t.Namespace,
				Name:              t.Name,
				CreationTimestamp: t.CreationTimestamp}}
	case *v1tf.TFJob:
		model = &CommonWorkload{
			commonWorkloadInfo: commonWorkloadInfo{
				key:          key,
				cacheIndexer: indexers[common.MedalType],
				labelKey:     labelKey,
				labelVal:     labelVal},
			containers: t.Spec.TFReplicaSpecs[v1common.ReplicaTypeWorker].Template.Spec.Containers,
			replicas:   *t.Spec.TFReplicaSpecs[v1common.ReplicaTypeWorker].Replicas,
			CommonPodInfo: agent2.CommonPodInfo{
				Namespace:         t.Namespace,
				Name:              t.Name,
				CreationTimestamp: t.CreationTimestamp}}

	default:
		return nil, fmt.Errorf("job factory err, %s ", key)
	}

	return model, nil
}

// RanktableFactory : return the version type of ranktable according to your input parameters
func RanktableFactory(model ResourceEventHandler, jobStartString, JSONVersion string) (v1.RankTabler, int32, error) {
	var ranktable v1.RankTabler
	groupList, replicasTotal, err := model.GenerateGrouplist()
	if err != nil {
		return nil, 0, fmt.Errorf("generate group list from job error: %v", err)
	}
	var rst v1.RankTableStatus
	err = rst.UnmarshalToRankTable(jobStartString)
	if err != nil {
		return nil, 0, err
	}
	if JSONVersion == "v1" {
		ranktable = &v1.RankTable{RankTableStatus: v1.RankTableStatus{Status: rst.Status},
			GroupCount: model.GetReplicas(), GroupList: groupList}
	} else {
		ranktable = &v2.RankTable{ServerCount: "0", ServerList: []*v2.Server(nil),
			RankTableStatus: v1.RankTableStatus{Status: rst.Status}, Version: "1.0"}
	}
	return ranktable, replicasTotal, nil
}
