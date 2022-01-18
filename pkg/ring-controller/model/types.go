/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package model

import (
	"hccl-controller/pkg/ring-controller/agent"
	apiCorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"time"
	v1alpha1apis "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

const (
	// VCJobType To determine the type of listening：vcjob.
	VCJobType = "vcjob"
	// DeploymentType To determine the type of listening：deployment.
	DeploymentType = "deployment"

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
