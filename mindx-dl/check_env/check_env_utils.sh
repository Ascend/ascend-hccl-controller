#!/usr/bin/env bash

# 常量
ONE_BLANK_SPACE_STR=" "
STATUS_NORMAL="Normal"
STATUS_ERROR="Error"
STATUS_INSTALLED="Installed"
STATUS_SERVER_NOT_INSTALL="Not install"
STATUS_POD_COMPLETED="Completed"
STATUS_PERMISSION_DENIED="Permission denied"
POD_STATUS_RUNNING="Running"
SERVICE_STATUS_RUNNING="active (running)"
NO_IMAGE="No image"
LINK_STATUS_DOWN="DOWN"

# 节点类型常量
MASTER_NODE="master"
WORKER_NODE="worker"
# 既是master又是worker
MASTER_WORKER_NODE="master-worker"

# 硬件形态常量
HW_COMMON="common"
HW_300T="300T"
HW_TRAIN="train"
HW_INFER="infer"

# 命令
UNSET_PROXY="unset http_proxy https_proxy"
GET_ALL_PODS_COMMAND="kubectl get pods --all-namespaces -o wide"
DESCRIBE_POD_COMMAND="kubectl describe pod"
DOCKER_IMAGES="docker images"

# 指定格式为：service*|*status*|*version*|*message code*|
# 说明：
#     一共四列，如果某一列的值为空，则使用一个空白字符串（如：" "）填充，效果为|* *|
#     最后会通过命令：column -s '*' -t 来格式化整个文件，达到对齐的目的

# 输出category
function write_category() {
    category="$1"

    local content="${category}*|* *|* *|* *|* *|"
    echo -e "${content}" >> "${tmp_output_file}"
}

# 构造指定格式的输出
function write_single_line_to_file() {
    local tmp_output_file="$1"
    local service="$2"
    local status="$3"
    local version="$4"
    local message_code="$5"

    if [[ "" == "${service}" ]]
    then
        service="${ONE_BLANK_SPACE_STR}"
    fi
    if [[ "" == "${status}" ]]
    then
        status="${ONE_BLANK_SPACE_STR}"
    fi
    if [[ "" == "${version}" ]]
    then
        version="${ONE_BLANK_SPACE_STR}"
    fi
    if [[ "" == "${message_code}" ]]
    then
        message_code="${ONE_BLANK_SPACE_STR}"
    fi

    local content=" *|*${service}*|*${status}*|*${version}*|*${message_code}*|"
    echo -e "${content}" >> "${tmp_output_file}"
}

# 已经构造好指定格式，直接写入
function write_multiline_to_file() {
    local tmp_output_file="$1"
    local content="$2"
    echo -e "${content}" >> "${tmp_output_file}"
}