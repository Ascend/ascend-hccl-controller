#!/bin/bash
# Perform  test for  hccl-controller
# Copyright (c) Huawei Technologies Co., Ltd. 2020-2023. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================
set -e

# execute go test and echo result to report files
function execute_test() {
  if ! (go test -mod=mod -v -race -coverprofile cov.out "${TOP_DIR}"/pkg/ring-controller/... >./"$file_input"); then
    echo '****** go test cases error! ******'
    exit 1
  else
    gocov convert cov.out | gocov-html >"$file_detail_output"
    gotestsum --junitfile unit-tests.xml "${TOP_DIR}"/pkg/ring-controller/...
    exit 0
  fi
}

function main() {
    echo "************************************* Start LLT Test *************************************"
    execute_test
    echo "************************************* End   LLT Test *************************************"
}
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
unset GOPATH
# if didn't install the following  tools, please install firstly
#go get -insecure github.com/axw/gocov/gocov
#go get github.com/matm/gocov-html
#go get github.com/golang/mock/mockgen
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)


file_input='testHccl.txt'
file_detail_output='api.html'

if [ -f "${TOP_DIR}"/test ]; then
  rm -rf "${TOP_DIR}"/test
fi
mkdir -p "${TOP_DIR}"/test
cd "${TOP_DIR}"/test
echo "clean old version test results"

if [ -f "$file_input" ]; then
  rm -rf "$file_input"
fi
if [ -f "$file_detail_output" ]; then
  rm -rf "$file_detail_output"
fi

main

