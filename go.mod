module hccl-controller

go 1.14

require (
	github.com/golang/mock v1.4.1
	github.com/prashantv/gostub v1.0.1-0.20191007164320-bbe3712b9c4a
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/klog v1.0.0
	volcano.sh/volcano v0.4.0
)

replace (
	k8s.io/api v0.0.0 => k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.17.8
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.17.8
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.17.8
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.17.8
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.17.8
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.17.8
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.17.8
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.17.8
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.17.8
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.17.8
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.17.8
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.17.8
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.17.8
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.17.8
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.17.8
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.17.8
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.17.8
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.17.8
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.17.8
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.17.8
)
