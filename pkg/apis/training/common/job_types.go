// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// JobStatus represents the current observed state of the training Job.
type JobStatus struct {
	// Conditions is an array of current observed job conditions.
	// +optional
	Conditions []JobCondition `json:"conditions"`

	// ReplicaStatuses is map of ReplicaType and ReplicaStatus,
	// specifies the status of each replica.
	// +optional
	ReplicaStatuses map[ReplicaType]*ReplicaStatus `json:"replicaStatuses"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Represents last time when the job was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

// +k8s:openapi-gen=true
// ReplicaType represents the type of the replica. Each operator needs to define its
// own set of ReplicaTypes.
type ReplicaType string

const (
	ReplicaTypeAM     ReplicaType = "AM"
	ReplicaTypeWorker ReplicaType = "Worker"
	ReplicaTypeBoard  ReplicaType = "Board"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// ReplicaSpec is a description of the replica
type ReplicaSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	Replicas *int32 `json:"replicas,omitempty"`

	// max attempts of replica type
	// if unspecified, defaults to 1.
	MaxAttempts *int32 `json:"maxAttempts,omitempty"`

	MarkingJobFinished bool `json:"markingJobFinished,omitempty"`

	// Template is the object that describes the pod that
	// will be created for this replica. RestartPolicy in PodTemplateSpec
	// will be overide by RestartPolicy in ReplicaSpec
	Template v1.PodTemplateSpec `json:"template,omitempty"`

	// Restart policy for all replicas within the job.
	// One of Always, OnFailure, Never and ExitCode.
	// Default to Never.
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty"`

	// +optional
	// DependOn represents a list of upstream vertex conditions to be dependent on for this RepicaType to start.
	// For example, in TensorFlow workers depend on ps to start first. If not set, KubeDL will populates the
	// default DependOn based on each framework's requirements. This feature is enabled by default, and can be
	// disabled with DAGScheduling feature gate.
	DependOn []DAGCondition `json:"-"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// JobCondition describes the state of the job at a certain point.
type JobCondition struct {
	// Type of job condition.
	Type JobConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +k8s:openapi-gen=true
// JobConditionType defines all kinds of types of JobStatus.
type JobConditionType string

const (
	// JobCreated means the job has been accepted by the system,
	// but one or more of the pods/services has not been started.
	// This includes time before pods being scheduled and launched.
	JobCreated JobConditionType = "Created"

	// JobRunning means all sub-resources (e.g. services/pods) of this job
	// have been successfully scheduled and launched.
	// The training is running without error.
	JobRunning JobConditionType = "Running"

	// JobSucceeded means all sub-resources (e.g. services/pods) of this job
	// reached phase have terminated in success.
	// The training is complete without error.
	JobSucceeded JobConditionType = "Succeeded"

	// JobFailed means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed with no restarting.
	// The training has failed its execution.
	JobFailed JobConditionType = "Failed"

	JobKilled JobConditionType = "Killed"
)

// +k8s:openapi-gen=true
// CleanPodPolicy describes how to deal with pods when the job is finished.
type CleanPodPolicy string

const (
	CleanPodPolicyUndefined   CleanPodPolicy = ""
	CleanPodPolicyAll         CleanPodPolicy = "All"
	CleanPodPolicyRunning     CleanPodPolicy = "Running"
	CleanPodPolicyNone        CleanPodPolicy = "None"
	CleanPodPolicyAllExceptAM CleanPodPolicy = "AllExceptAM"
)

// +k8s:openapi-gen=true
// RestartPolicy describes how the replicas should be restarted.
// Only one of the following restart policies may be specified.
// If none of the following policies is specified, the default one
// is RestartPolicyAlways.
type RestartPolicy string

const (
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	// RestartPolicyExitCode policy means that user should add exit code by themselves,
	// The job operator will check these exit codes to
	// determine the behavior when an error occurs:
	// - 1-127: permanent error, do not restart.
	// - 128-255: retryable error, will restart the pod.
	RestartPolicyExitCode RestartPolicy = "ExitCode"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// SchedulingPolicy define the scheduling related params for a job
type SchedulingPolicy struct {
	// enable gang scheduling policy, default to false
	// +optional
	EnableGangScheduling bool `json:"enableGangScheduling,omitempty"`
	// priority of the job's podgroup
	// +optional
	JobPriority *int32 `json:"jobPriority,omitempty"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// JobMetaInfo define the meta info of training job
type JobMetaInfo struct {
	MisId   string `json:"misId,omitempty"`
	JobName string `json:"jobName,omitempty"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// RunPolicy encapsulates various runtime policies of the distributed training
// job, for example how to clean up resources and how long the job can stay
// active.
type RunPolicy struct {
	// CleanPodPolicy defines the policy to kill pods after the job completes.
	// Default to Running.
	// +optional
	CleanPodPolicy *CleanPodPolicy `json:"cleanPodPolicy,omitempty"`

	// TTLSecondsAfterFinished is the TTL to clean up jobs.
	// It may take extra ReconcilePeriod seconds for the cleanup, since
	// reconcile gets called periodically.
	// Default to infinite.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// Specifies the duration in seconds relative to the startTime that the job may be active
	// before the system tries to terminate it; value must be positive integer.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Optional number of retries before marking this job failed.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Job schedule timeout seconds
	// +optional
	ScheduleTimeoutSecond *int32 `json:"scheduleTimeoutSecond,omitempty"`

	// Support supportFailover
	// +optional
	SupportFailover bool `json:"supportFailover,omitempty"`
}

type SuccessPolicy string

const (
	SuccessPolicyDefault    SuccessPolicy = ""
	SuccessPolicyAllWorkers SuccessPolicy = "AllWorkers"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
type DAGCondition struct {
	// Upstream defines which replica type is the source tigger.
	Upstream ReplicaType `json:"upstream"`
	// OnPhase defines at which phase the upstream replica will trigger this condition.
	OnPhase ReplicaConditionType `json:"onPhase"`

	Worker0Last bool `json:"worker0Last"`
}
