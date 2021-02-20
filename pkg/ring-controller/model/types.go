/*
 * Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"hccl-controller/pkg/ring-controller/agent"
	apiCorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"time"
	v1alpha1apis "volcano.sh/volcano/pkg/apis/batch/v1alpha1"
)

const (

	// L2 log level 2
	L2 = 2
	// L3 log level 3
	L3 = 3
	// L4 log level 4
	L4 = 4

	decimal = 10
	// VCJobType To determine the type of listening：vcjob.
	VCJobType = "vcjob"
	// DeploymentType To determine the type of listening：deployment.
	DeploymentType = "deployment"
	// JobRestartPhase restart flage
	JobRestartPhase = "Restarting"

	// BuildStatInterval 1 * time.Minute
	BuildStatInterval = 30 * time.Second
)

type modelCommon struct {
	key          string
	cacheIndexer cache.Indexer
}

// VCJobModel : to handle vcjob type
type VCJobModel struct {
	modelCommon
	agent.JobInfo
	jobPhase string
	taskSpec []v1alpha1apis.TaskSpec
}

// DeployModel : to handle deployment type
type DeployModel struct {
	modelCommon
	agent.DeployInfo
	replicas   int32
	containers []apiCorev1.Container
}
