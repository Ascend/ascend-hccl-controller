#!/bin/bash
# Perform  build hccl-controller
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
export GO111MODULE="on"
unset GOPATH
build_version="v0.0.1"
build_time=$(date +'%Y-%m-%d_%T')
OUTPUT_NAME="hccl-controller"



DOCKER_FILE_NAME="Dockerfile"
docker_zip_name="hccl-controller.tar.gz"
docker_images_name="hccl-controller:${build_version}"
function clear_env()
{
    rm -rf ${TOP_DIR}/output/*
}

function build()
{
    cd ${TOP_DIR}
    go build -ldflags "-X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version} \
            -X main.BuildTime=${build_time}"  \
            -o ${OUTPUT_NAME}
    ls ${OUTPUT_NAME}
    if [ $? -ne 0 ]; then
        echo "fail to find hccl-controller"
        exit 1
    fi
}

function mv_file()
{
    mv ${TOP_DIR}/${OUTPUT_NAME}   ${TOP_DIR}/output
    mv ${TOP_DIR}/build/*.yaml    ${TOP_DIR}/output
}

function build_docker_image() {
    cp ${TOP_DIR}/build/${DOCKER_FILE_NAME}     ${TOP_DIR}/output
    cd ${TOP_DIR}/output
    docker build -t ${docker_images_name} .
    docker save ${docker_images_name} | gzip > ${docker_zip_name}
    rm -f ${DOCKER_FILE_NAME}
}

function main() {
    clear_env
    build
    mv_file
    build_docker_image
}

main