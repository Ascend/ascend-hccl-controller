module hccl-controller

go 1.14

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	huawei.com/mindx/common/hwlog v0.0.0
	huawei.com/mindx/common/k8stool v0.0.0
	huawei.com/mindx/common/utils v0.0.0
	k8s.io/api v0.19.11
	k8s.io/apimachinery v0.19.11
	k8s.io/client-go v0.19.11
	volcano.sh/apis v0.0.0-20210603070204-70005b2d502a
)

replace (
	github.com/agiledragon/gomonkey/v2 v2.0.1 => github.com/agiledragon/gomonkey/v2 v2.1.0
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.0.2
	huawei.com/mindx/common/k8stool => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/k8stool v0.0.2
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.0.5
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.0.2
	k8s.io/api v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/api v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/apimachinery v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/apimachinery v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/client-go v0.0.0 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
	k8s.io/client-go v0.19.11 => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
)
