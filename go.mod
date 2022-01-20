module hccl-controller

go 1.14

require (
	github.com/agiledragon/gomonkey/v2 v2.1.0
	github.com/smartystreets/goconvey v1.6.4
	huawei.com/npu-exporter v0.1.7
	k8s.io/api v0.19.11
	k8s.io/apimachinery v0.19.11
	k8s.io/client-go v0.19.11
	volcano.sh/apis v0.0.0-20210603070204-70005b2d502a
)

replace (
	github.com/agiledragon/gomonkey/v2 v2.0.1 => github.com/agiledragon/gomonkey/v2 v2.1.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	huawei.com/kmc => codehub-dg-y.huawei.com/it-edge-native/edge-native-core/coastguard.git v1.0.6
	huawei.com/npu-exporter => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/npu-exporter.git v0.1.7
	k8s.io/api v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/api v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/apimachinery v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/apiserver => k8s.io/apiserver v0.19.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.4
	k8s.io/client-go v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
	k8s.io/client-go v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.19.4
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.19.4
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.19.4
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.19.4
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.19.4
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.19.4
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.19.4
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.19.4
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.19.4
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.19.4
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.19.4
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.19.4
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.19.4
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.19.4
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.19.4
)
