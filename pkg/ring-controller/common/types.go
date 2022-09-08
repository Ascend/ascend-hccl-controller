package common

const (
	// DeploymentType To determine the type of listening：deployment.
	DeploymentType = "deployment"
	// K8sJobType To determine the type of listening：job.
	K8sJobType = "job"
	// MedalType To determine the type of listening：medaljob.
	MedalType = "medaljob"
	// MpiType To determine the type of listening：mpijob.
	MpiType = "mpijob"
	// TfType To determine the type of listening：tfjob.
	TfType = "tfjob"
	// ReplicaSetType ReplicaSet is the owner of deployment type pod
	ReplicaSetType = "replicaset"
)
