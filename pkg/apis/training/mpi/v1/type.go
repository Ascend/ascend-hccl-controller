package v1

import (
	commonv1 "hccl-controller/pkg/apis/training/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=mpijob
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="msg",type=string,JSONPath=`.status.conditions[-1:].message`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MPIJob struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec MPIJobSpec `json:"spec,omitempty"`
	// Read-only.
	// +optional
	Status commonv1.JobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=mpijobs
// +kubebuilder:object:root=true
type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MPIJob `json:"items"`
}

type MPIJobSpec struct {

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 1.
	// +optional
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`

	// RunPolicy encapsulates various runtime policies of the job.
	RunPolicy commonv1.RunPolicy `json:"runPolicy"`

	// SuccessPolicy defines the policy to mark the TFJob as succeeded.
	// Default to "", using the default rules.
	// +optional
	SuccessPolicy *commonv1.SuccessPolicy `json:"successPolicy,omitempty"`

	// MPIReplicaSpecs contains maps from `MPIReplicaType` to `ReplicaSpec` that
	// specify the MPI replicas to run.
	MPIReplicaSpecs map[commonv1.ReplicaType]*commonv1.ReplicaSpec `json:"mpiReplicaSpecs"`

	// SSHAuthMountPath is the directory where SSH keys are mounted.
	// Defaults to "/root/.ssh".
	SSHAuthMountPath string `json:"sshAuthMountPath,omitempty"`
}

const (
	MPIReplicaTypePS   commonv1.ReplicaType = "PS"
	MPIReplicaTypeEval commonv1.ReplicaType = "Evaluator"
)
