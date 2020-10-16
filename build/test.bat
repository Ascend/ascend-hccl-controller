rem Just for developer to generate mock directory and files
rem Need to install mockgen firstly go get github.com/golang/mock/mockgen
@echo off
cd /d %~dp0
mkdir %~dp0..\pkg\ring-controller\controller\mock_cache
mkdir %~dp0..\pkg\ring-controller\controller\mock_controller
mkdir %~dp0..\pkg\ring-controller\controller\mock_kubernetes
mkdir %~dp0..\pkg\ring-controller\controller\mock_v1
mkdir %~dp0..\pkg\ring-controller\controller\mock_v1alpha1


mockgen k8s.io/client-go/kubernetes/typed/core/v1 ConfigMapInterface  >%~dp0..\pkg\ring-controller\controller\\mock_v1\configMapInterface_mock.go

mockgen k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface  >%~dp0..\pkg\ring-controller\controller\\mock_v1\corev1_mock.go

mockgen k8s.io/client-go/tools/cache SharedIndexInformer  >%~dp0..\pkg\ring-controller\controller\mock_cache\sharedInformer_mock.go

mockgen k8s.io/client-go/tools/cache Indexer >%~dp0..\pkg\ring-controller\controller\mock_cache\indexer_mock.go

mockgen hccl-controller/pkg/ring-controller/controller WorkAgentInterface  >%~dp0..\pkg\ring-controller\controller\mock_controller\businessagent_mock.go

mockgen k8s.io/client-go/kubernetes Interface  >%~dp0..\pkg\ring-controller\controller\mock_kubernetes\k8s_interface_mock.go

mockgen volcano.sh/volcano/pkg/client/informers/externalversions/batch/v1alpha1 JobInformer  >%~dp0..\pkg\ring-controller\controller\mock_v1alpha1\informer_mock.go


go test -v -race -coverprofile cov.out  hccl-controller/pkg/ring-controller/controller

del  %~dp0\cov.out
rmdir /S/Q %~dp0..\pkg\ring-controller\controller\mock_cache
rmdir /S/Q %~dp0..\pkg\ring-controller\controller\mock_controller
rmdir /S/Q %~dp0..\pkg\ring-controller\controller\mock_kubernetes
rmdir /S/Q %~dp0..\pkg\ring-controller\controller\mock_v1
rmdir /S/Q %~dp0..\pkg\ring-controller\controller\mock_v1alpha1
@echo on
