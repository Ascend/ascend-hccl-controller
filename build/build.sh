#!/bin/bash
# Perform  build hccl-controller
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v2.0.2"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '3p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"
if [ "${arch:0:5}" = 'aarch' ]; then
  arch=arm64
else
  arch=amd64
fi

sed -i "s/hccl-controller:.*/hccl-controller:${build_version}/" "${TOP_DIR}"/build/hccl-controller.yaml

OUTPUT_NAME="hccl-controller"

DOCKER_FILE_NAME="Dockerfile"
docker_zip_name="hccl-controller-${build_version}-${arch}.tar.gz"
docker_images_name="hccl-controller:${build_version}"
function clear_env() {
  rm -rf "${TOP_DIR}"/output/*
}

function build() {
  cd "${TOP_DIR}"

  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_ENABLED=0 go build -mod=mod -buildmode=pie   -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now  -X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version}" \
            -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find hccl-controller"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}"/${OUTPUT_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/hccl-controller.yaml "${TOP_DIR}"/output/hccl-controller-"${build_version}".yaml
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
}

function change_mod() {
  chmod 400 "${TOP_DIR}"/output/*
  chmod 500 "${TOP_DIR}/output/${OUTPUT_NAME}"
}



function main() {
  clear_env
  build
  mv_file
  change_mod
}

main
