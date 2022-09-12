package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// TaskStatus represents the current observed state of the task.
type TaskAttemptStatus struct {
	PodName       string `json:"podName,omitempty"`
	PodUID        string `json:"podUID,omitempty"`
	TaskId        int32  `json:"taskId,omitempty"`
	TaskAttemptId int32  `json:"taskAttemptId,omitempty"`
	// +optional
	NodeName string `json:"nodeName,omitempty"`
	// +optional
	PodIP string `json:"podIP,omitempty"`
	// +optional
	Conditions []TaskAttemptCondition `json:"conditions"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// TaskAttemptCondition describes the state of the task at a certain point.
type TaskAttemptCondition struct {
	Type               TaskAttemptConditionType `json:"type"`
	Status             v1.ConditionStatus       `json:"status"`
	Reason             string                   `json:"reason,omitempty"`
	Message            string                   `json:"message,omitempty"`
	LastUpdateTime     metav1.Time              `json:"lastUpdateTime,omitempty"`
	LastTransitionTime metav1.Time              `json:"lastTransitionTime,omitempty"`
}

type PodChange struct {
	Role          string               `json:"role,omitempty"`
	TaskId        int32                `json:"taskId,omitempty"`
	TaskAttemptId int32                `json:"taskAttemptId,omitempty"`
	Condition     TaskAttemptCondition `json:"condition"`
}

// +k8s:openapi-gen=true
// TaskAttemptConditionType defines all kinds of types of TaskAttemptStatus.
type TaskAttemptConditionType string

const (
	TaskAttemptCreated TaskAttemptConditionType = "Pending"

	TaskAttemptRunning TaskAttemptConditionType = "Running"

	TaskAttemptSucceeded TaskAttemptConditionType = "Succeeded"

	TaskAttemptFailed TaskAttemptConditionType = "Failed"

	TaskAttemptUnknown TaskAttemptConditionType = "Unknown"

	TaskAttemptFinished TaskAttemptConditionType = "Finished"
)
