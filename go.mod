module hccl-controller

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/smartystreets/goconvey v1.6.4
	go.uber.org/zap v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)

replace (
	github.com/agiledragon/gomonkey/v2 v2.0.1 => github.com/agiledragon/gomonkey/v2 v2.1.0
	k8s.io/api v0.0.0 => k8s.io/api v0.21.2
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.21.2
)
