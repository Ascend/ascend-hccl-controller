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
	"time"

	apiCorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"hccl-controller/pkg/ring-controller/agent"
)

const (
	decimal = 10

	// BuildStatInterval 1 * time.Minute
	BuildStatInterval = 30 * time.Second
)

type commonWorkloadInfo struct {
	key          string
	labelKey     string
	labelVal     string
	cacheIndexer cache.Indexer
}

// CommonWorkload : to handle deployment, job workloads type
type CommonWorkload struct {
	commonWorkloadInfo
	agent.CommonPodInfo
	replicas   int32
	containers []apiCorev1.Container
}
