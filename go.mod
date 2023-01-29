module hccl-controller

go 1.14

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	huawei.com/npu-exporter/v3 v3.0.0
	k8s.io/api v0.19.11
	k8s.io/apimachinery v0.19.11
	k8s.io/client-go v0.19.11
	volcano.sh/apis v0.0.0-20210603070204-70005b2d502a
)

replace (
	github.com/agiledragon/gomonkey/v2 v2.0.1 => github.com/agiledragon/gomonkey/v2 v2.1.0
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.1
	huawei.com/npu-exporter/v3 => gitee.com/ascend/ascend-npu-exporter/v3 v3.0.0
	k8s.io/api v0.19.11 => k8s.io/api v0.19.11
	k8s.io/apimachinery v0.19.11 => k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.11 => k8s.io/client-go v0.19.4
)
