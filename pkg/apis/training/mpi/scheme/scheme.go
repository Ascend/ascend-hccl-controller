package scheme

import (
	v1 "hccl-controller/pkg/apis/training/mpi/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()

	// Codecs provides access to encoding and decoding for the scheme
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	Install(Scheme)
}

// Install registers the API group and adds types to a scheme.
func Install(scheme *runtime.Scheme) {
	v1.AddToScheme(scheme)
}
