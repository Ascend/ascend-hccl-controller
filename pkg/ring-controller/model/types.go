/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 *
 */

package model

import (
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"hccl-controller/pkg/ring-controller/agent"
)

const (
	// VCJobType To determine the type of listening：vcjob.
	VCJobType = "vcjob"
	// DeploymentType To determine the type of listening：deployment.
	DeploymentType = "deployment"

	// BuildStatInterval 30 * time.Second
	BuildStatInterval = 30 * time.Second

	maxContainerNum = 2
	maxNodeNum      = 256
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
	taskSpec []v1alpha1.TaskSpec
}

// DeployModel : to handle deployment type
type DeployModel struct {
	modelCommon
	agent.DeployInfo
	replicas   int32
	containers []v1.Container
}
