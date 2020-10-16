#!/bin/bash
# Perform  test for  hccl-controller
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
unset GOPATH
# if didn't install the following  tools, please install firstly
#go get -insecure github.com/axw/gocov/gocov
#go get github.com/matm/gocov-html
#go get github.com/golang/mock/mockgen
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
MOCK_TOP=${TOP_DIR}/pkg/ring-controller/controller
echo "start generate mock files"
mkdir -p "${TOP_DIR}"/pkg/ring-controller/controller/mock_cache
mkdir -p "${TOP_DIR}"/pkg/ring-controller/controller/mock_controller
mkdir -p "${TOP_DIR}"/pkg/ring-controller/controller/mock_kubernetes
mkdir -p "${TOP_DIR}"/pkg/ring-controller/controller/mock_v1
mkdir -p "${TOP_DIR}"/pkg/ring-controller/controller/mock_v1alpha1

mockgen k8s.io/client-go/kubernetes/typed/core/v1 ConfigMapInterface > "${MOCK_TOP}"/mock_v1/configMapInterface_mock.go
mockgen k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface > "${MOCK_TOP}"/mock_v1/corev1_mock.go
mockgen volcano.sh/volcano/pkg/client/informers/externalversions/batch/v1alpha1 JobInformer >  "${MOCK_TOP}"/mock_v1alpha1/former_mock.go
mockgen k8s.io/client-go/kubernetes Interface >  "${MOCK_TOP}"/mock_kubernetes/k8s_interface_mock.go
mockgen hccl-controller/pkg/ring-controller/controller WorkAgentInterface > "${MOCK_TOP}"/mock_controller/businessagent_mock.go
mockgen k8s.io/client-go/tools/cache Indexer > "${MOCK_TOP}"/mock_cache/indexer_mock.go
mockgen k8s.io/client-go/tools/cache SharedIndexInformer >  "${MOCK_TOP}"/mock_cache/sharedInformer_mock.go
echo "finish generate mock files"

file_input='testHccl.txt'
file_detail_output='hcclCoverageReport.html'

cd  "${TOP_DIR}"/output
echo "clean old version  test results"
if  [ -f "$file_input" ]; then
  rm -rf "$file_input"
fi
if  [ -f "$file_detail_output" ]; then
  rm -rf "$file_detail_output"
fi

echo "************************************* Start LLT Test *************************************"
if ! (go test -v -race -coverprofile cov.out "${TOP_DIR}"/pkg/ring-controller/controller > ./$file_input)
then
  echo '****** go test cases error! ******'
  echo 'Failed' > $file_input
else
  gocov convert cov.out | gocov-html > $file_detail_output
fi

{
  echo "<html<body><h1>==================================================</h1><table border='2'>"
  echo "<html<body><h1>HCCL testCase</h1><table border='1'>"
  echo "<html<body><h1>==================================================</h1><table border='2'>"
} >> ./$file_detail_output

while read -r line
do
  echo -e "<tr>
     $(echo "$line" | awk 'BEGIN{FS="|"}''{i=1;while(i<=NF) {print "<td>"$i"</td>";i++}}')
  </tr>" >> $file_detail_output
done < $file_input
echo "</table></body></html>" >> ./$file_detail_output

echo "************************************* End   LLT Test *************************************"
echo "start clean mock files"
rm -rf "${MOCK_TOP}"/mock_cache
rm -rf "${MOCK_TOP}"/mock_controller
rm -rf "${MOCK_TOP}"/mock_kubernetes
rm -rf "${MOCK_TOP}"/mock_v1
rm -rf "${MOCK_TOP}"/mock_v1alpha1
echo "finish clean mock files"
