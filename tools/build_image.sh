#!/bin/bash

APP_PATH=""
APP=""
TAG=""
ARCH=""
HARBOR_HOST=""
HARBOR_PORT=""
YML_PATH=""
YML_FILE=""
YML_DIR=""
IMAGE=""
DEBUG="n"
DRYRUN="n"
PROJECT="mindx"
LOCAL_ARCH=$(uname -m)
DL_COMPONENTS="npu-exporter,device-plugin"

function print_usage()
{
    echo "Usage:"
    echo "$0 [options] zip_file (only for ${DL_COMPONENTS})"
    echo "options:"
    echo "    --help         print this message"
    echo "    --debug        enable debug"
    echo "    --dryrun       not running, just test process"
    echo "    --harbor-ip    set harbor ip"
    echo "    --harbor-port  set harbor port, default 7443"
}

function parse_args()
{
    if [ $# = 0 ];then
        print_usage
        exit 1
    fi
    while true; do
        case "$1" in
        --help | -h)
            print_usage
            exit 0
            ;;
        --debug)
            DEBUG="y"
            shift
            ;;
        --dryrun)
            DRYRUN="y"
            shift
            ;;
        --harbor-ip)
            HARBOR_HOST=$2
            shift
            shift
            ;;
        --harbor-port)
            HARBOR_PORT=":$2"
            shift
            shift
            ;;
        *)
            if [ "x$1" != "x" ]; then
                APP_PATH=$1
                return 0
            fi
            break
            ;;
        esac
    done
    if [ $DEBUG == "y" ]; then
        echo "DEBUG: DEBUG=$DEBUG"
        echo "DEBUG: DRYRUN=$DRYRUN"
        echo "DEBUG: APP_PATH=$APP_PATH"
        echo "DEBUG: HARBOR_HOST=$HARBOR_HOST"
        echo "DEBUG: HARBOR_PORT=$HARBOR_PORT"
    fi
}

function get_app_info()
{
    unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY
    if [[ ${APP_PATH} =~ "device-plugin" ]];then
        DEVICE_YAMLS=()
        if [[ $(kubectl get node -l accelerator=huawei-Ascend310 | wc -l) != 0 ]];then
            DEVICE_YAMLS[${#DEVICE_YAMLS[@]}]=$(find ${APP_PATH} -maxdepth 1 -name '*-v*.yaml' | grep "310-volcano")
        fi
        if [[ $(kubectl get node -l accelerator=huawei-Ascend310P | wc -l) != 0 ]];then
            DEVICE_YAMLS[${#DEVICE_YAMLS[@]}]=$(find ${APP_PATH} -maxdepth 1 -name '*-v*.yaml' | grep "310P-volcano")
        fi
        if [[ $(kubectl get node -l accelerator=huawei-Ascend910 | wc -l) != 0 ]];then
            DEVICE_YAMLS[${#DEVICE_YAMLS[@]}]=$(find ${APP_PATH} -maxdepth 1 -name '*-v*.yaml' | grep "device-plugin-volcano")
        fi
        if [[ ${#DEVICE_YAMLS[@]} == 0 ]]; then
            echo "FATAL: can not find npu card, device-plugin install failed"
            exit 1
        fi
        YML_PATH=${DEVICE_YAMLS[0]}
        echo "device-plugin yaml count: ${#DEVICE_YAMLS[@]}"
    else
        YML_PATH=$(find ${APP_PATH} -maxdepth 1 -name '*-v*.yaml' | grep -v '\-without-token-')
    fi
    if [ -z "$YML_PATH" ];then
        echo "FATAL: no yaml file found, please check ${APP_PATH}"
        exit 1
    fi
    YML_DIR=$(dirname $YML_PATH)
    YML_FILE=$(basename $YML_PATH)
    APP=$(echo $YML_FILE | cut -d"-" -f1)
    TAG=$(grep "image: " $YML_PATH  | head -n1 | cut -d":" -f3)
    local tmp=$(grep "image: " $YML_PATH  | head -n1 | cut -d":" -f2)
    IMAGE=$(echo ${tmp} | xargs echo -n)
    ## get executable arch
    local exe_file=$(find ${APP_PATH} -maxdepth 1 -type f -executable | head -n1)
    if [ -z ${exe_file} ];then
        exe_file=$(find ${APP_PATH}/lib/*.so -maxdepth 1 -type f | head -n1)
    fi
    local arch_flag=$(file ${exe_file} | grep x86 | wc -l)
    if [ ${arch_flag} == "1" ];then
        ARCH="amd64"
    else
        ARCH="arm64"
    fi
    echo "YML_FILE=${YML_FILE}"
    echo "APP=${APP}"
    echo "TAG=${TAG}"
    echo "ARCH=${ARCH}"
    local arch_check=$(file ${exe_file} | grep ${LOCAL_ARCH} | wc -l)
    if [ ${arch_check} == "0" ];then
        echo "FATAL: cpu arch is ${LOCAL_ARCH} but package is ${ARCH}"
        exit 1
    fi
}

function modify_dockerfile()
{
    local docker_file=$1
    local from_line=$(grep "FROM " ${docker_file})
    local base_img=$(echo ${from_line} | cut -d " " -f2)
    local image=$(echo ${base_img} | cut -d ":" -f1)
    local tag=$(echo ${base_img} | cut -d ":" -f2)
    if [[ "${base_img}" == "${HARBOR_HOST}"* ]];then
        return
    fi
    if [[ "${base_img}" == "ubuntu"* ]];then
        sed -i "s#^FROM.*#FROM ${HARBOR_HOST}${HARBOR_PORT}/dockerhub/${image}_${ARCH}:${tag}#g" ${docker_file}
    fi
}

function build_image()
{
    local docker_file=$1
    local name=$2
    local arch=$3
    local tag=$4

    if [ ${DEBUG} == "y" ];then
        echo "DEBUG: docker build . -f ${docker_file} -t ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag}"
        echo "DEBUG: docker push ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag}"
        echo "DEBUG: docker manifest create ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag} -a"
        echo "DEBUG: docker manifest annotate ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag} --os linux --arch ${arch}"
        echo "DEBUG: docker manifest push ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} -p"
    fi

    if [ ${DRYRUN} == "y" ];then
        return
    fi

    docker build . -f ${docker_file} -t ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag}
    if [ $? != 0 ];then
        echo "docker build failed"
        exit 1
    fi
    docker push ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag}
    if [ $? != 0 ];then
        echo "push to harbor failed"
        exit 1
    fi
    docker manifest create ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag} -a
    if [ $? != 0 ];then
        echo "create manifest failed"
        exit 1
    fi
    docker manifest annotate ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}_${arch}:${tag} --os linux --arch ${arch}
    if [ $? != 0 ];then
        echo "manifest annotate failed"
        exit 1
    fi
    docker manifest push ${HARBOR_HOST}${HARBOR_PORT}/${PROJECT}/${name}:${tag} -p
    if [ $? != 0 ];then
        echo "push manifest failed"
        exit 1
    fi
    cd -
}

function build_images()
{
    cd ${APP_PATH}
    for docker_file in `find . -maxdepth 1 -type f -name "Docker*"`
    do
        modify_dockerfile ${docker_file}
        local file_name=$(basename $docker_file)
        echo "build image for $file_name:"
        if [ ${file_name} == "Dockerfile" ];then
            build_image ${docker_file} ${IMAGE} ${ARCH} ${TAG}
        fi
    done
}

function do_install()
{
    export DOCKER_CLI_EXPERIMENTAL=enabled
    build_images
}

function check_components()
{
    IFS=","
    for component in $1
    do
        if [[ "$2" =~ ${component} ]];then
            echo 1
            unset IFS
            return 0
        fi
    done
    echo 0
    unset IFS
    return 0
}

function unarchive_app()
{
    local base_dir=$(dirname $APP_PATH)
    local tmp=${APP_PATH%.zip}
    local dir_name=${tmp##*/}
    echo "unarchive to ${base_dir}/${dir_name}"
    rm -rf ${base_dir}/${dir_name}
    unzip ${APP_PATH} -d ${base_dir}/${dir_name}
    APP_PATH=$base_dir/$dir_name
}

function pre_check()
{
    local have_docker=$(command -v docker | wc -l)
    if [ ${have_docker} -eq 0 ]; then
        echo "can not find docker"
        exit 1
    fi
    local have_kubectl=$(command -v kubectl | wc -l)
    if [ ${have_kubectl} -eq 0 ]; then
        echo "can not find kubectl"
        exit 1
    fi
    if [[ -z ${APP_PATH} ]] || [[ ! -f ${APP_PATH} ]];then
        echo "expect a valid path"
        exit 1
    fi
    if [ -z ${HARBOR_HOST} ];then
        echo "no harbor ip set, please set harbor by --harbor-ip"
        exit 1
    fi
    if [ -z ${HARBOR_PORT} ];then
        echo "no harbor port set, please set harbor by --harbor-port"
        exit 1
    fi
    if [[ $(check_components ${DL_COMPONENTS} ${APP_PATH}) == 0 ]]; then
        echo "only support build images for ${DL_COMPONENTS}"
        exit 1
    fi
}

function main()
{
    parse_args $*
    pre_check
    if [ -f ${APP_PATH} ];then
        unarchive_app
    fi
    get_app_info
    do_install
}

main $*
