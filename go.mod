module hccl-controller

go 1.14

require (
	github.com/agiledragon/gomonkey/v2 v2.0.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/smartystreets/goconvey v1.6.4
	go.uber.org/zap v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	volcano.sh/volcano v0.4.0
)

replace (
	github.com/agiledragon/gomonkey/v2 v2.0.1 => github.com/agiledragon/gomonkey/v2 v2.1.0
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
