package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// TaskStatus represents the current observed state of the task.
type TaskStatus struct {
	// +optional
	TaskId int32 `json:"taskId,omitempty"`
	// +optional
	Conditions []TaskCondition `json:"conditions"`
	// +optional
	TaskAttemptStatuses []TaskAttemptStatus `json:"taskAttemptStatuses"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// TaskCondition describes the state of the task at a certain point.
type TaskCondition struct {
	Type               TaskConditionType  `json:"type"`
	Status             v1.ConditionStatus `json:"status"`
	Reason             string             `json:"reason,omitempty"`
	Message            string             `json:"message,omitempty"`
	LastUpdateTime     metav1.Time        `json:"lastUpdateTime,omitempty"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime,omitempty"`
}

// +k8s:openapi-gen=true
// TaskConditionType defines all kinds of types of TaskStatus.
type TaskConditionType string

const (
	TaskCreated TaskConditionType = "Created"

	TaskRunning TaskConditionType = "Running"

	TaskSucceeded TaskConditionType = "Succeeded"

	TaskFailed TaskConditionType = "Failed"

	TaskFinished TaskConditionType = "Finished"
)
