#!/bin/bash
# Perform  build hccl-controller
# Copyright (c) Huawei Technologies Co., Ltd. 2020-2022. All rights reserved.
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
cur_dir=$(dirname "$(readlink -f "$0")")
top_dir=$(realpath "${cur_dir}"/..)
export GO111MODULE="on"
ver_file="${top_dir}"/service_config.ini
build_version="v5.0.RC1"
if [ -f "$ver_file" ]; then
  line=$(sed -n '1p' "$ver_file" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi
arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

output_name="hccl-controller"

docker_file_name="Dockerfile"
docker_zip_name="hccl-controller-${build_version}-${arch}.tar.gz"
docker_images_name="hccl-controller:${build_version}"
function clean() {
  rm -rf "${top_dir}"/output/*
  mkdir -p "${top_dir}"/output
}

function build() {
  cd "${top_dir}"

  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now  -X main.BuildName=${output_name} \
            -X main.BuildVersion=${build_version}_linux-${arch}" \
  -o ${output_name}
  ls ${output_name}
  if [ $? -ne 0 ]; then
    echo "fail to find hccl-controller"
    exit 1
  fi
}

function mv_file() {
  mv "${top_dir}"/${output_name} "${top_dir}"/output
  cp "${top_dir}"/build/hccl-controller.yaml "${top_dir}"/output/hccl-controller-"${build_version}".yaml
  sed -i "s/hccl-controller:.*/hccl-controller:${build_version}/" \
  "${top_dir}"/output/hccl-controller-"${build_version}".yaml

  cp "${top_dir}"/build/${docker_file_name} "${top_dir}"/output
}

function change_mod() {
  chmod 400 "${top_dir}"/output/*
  chmod 500 "${top_dir}/output/${output_name}"
}

function main() {
  clean
  build
  mv_file
  change_mod
}

if [ "$1" = clean ]; then
  clean
  exit 0
fi

main $1
