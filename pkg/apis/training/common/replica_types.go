package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// ReplicaStatus represents the current observed state of the replica.
type ReplicaStatus struct {
	// +optional
	Conditions []ReplicaCondition `json:"conditions"`
	// +optional
	TaskStatuses []TaskStatus `json:"taskStatuses"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=true
// ReplicaCondition describes the state of the replica at a certain point.
type ReplicaCondition struct {
	Type               ReplicaConditionType `json:"type"`
	Status             v1.ConditionStatus   `json:"status"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
	LastUpdateTime     metav1.Time          `json:"lastUpdateTime,omitempty"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
}

// +k8s:openapi-gen=true
// ReplicaConditionType defines all kinds of types of ReplicaStatus.
type ReplicaConditionType string

const (
	ReplicaCreated ReplicaConditionType = "Created"

	ReplicaRunning ReplicaConditionType = "Running"

	ReplicaSucceeded ReplicaConditionType = "Succeeded"

	ReplicaFailed ReplicaConditionType = "Failed"

	ReplicaFinished ReplicaConditionType = "Finished"
)
