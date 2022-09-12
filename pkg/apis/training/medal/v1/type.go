package v1

import (
	commonv1 "hccl-controller/pkg/apis/training/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=medaljob
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="msg",type=string,JSONPath=`.status.conditions[-1:].message`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MedalJob struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec MedalJobSpec `json:"spec,omitempty"`
	// Read-only.
	// +optional
	Status commonv1.JobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=medaljobs
// +kubebuilder:object:root=true
type MedalJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MedalJob `json:"items"`
}

type MedalJobSpec struct {

	// RunPolicy encapsulates various runtime policies of the job.
	RunPolicy commonv1.RunPolicy `json:"runPolicy"`

	// SchedulingPolicy defines the policy of job schedule
	// +optional
	SchedulingPolicy commonv1.SchedulingPolicy `json:"schedulingPolicy"`

	// +optional
	JobMetaInfo commonv1.JobMetaInfo `json:"jobMetaInfo"`

	// SuccessPolicy defines the policy to mark the TFJob as succeeded.
	// Default to "", using the default rules.
	// +optional
	SuccessPolicy *commonv1.SuccessPolicy `json:"successPolicy,omitempty"`

	// MedalReplicaSpecs contains maps from `MedalReplicaSpecs` to `ReplicaSpec` that
	// specify the Medal replicas to run.
	MedalReplicaSpecs map[commonv1.ReplicaType]*commonv1.ReplicaSpec `json:"medalReplicaSpecs"`
}

const (
	MedalReplicaTypeEM commonv1.ReplicaType = "EM"
)
