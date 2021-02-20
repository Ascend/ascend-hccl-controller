#!/usr/bin/env bash

# 输入参数
nodeType="$1"
# hardWare="$2"
# host_ip="$3"
tmp_output_file="$4"

# 消息码
# error类
ERROR_DOCKER_NOT_INSTALL_CODE="error_3001"
ERROR_DOCKER_SERVICE_CODE="error_3002"
ERROR_DOCKER_COMMAND_CODE="error_3003"
# warning类
WARNING_NOT_ASCEND_DOCKER_CODE="warning_3001"
WARNING_IMAGE_DISK_USAGE_CODE="warning_3002"
# info类
INFO_DOCKER_VERSION_RECOMMEND_CODE="info_3001"

docker_info="$(docker info 2>/dev/null)"
ascend_runtime="ascend"
DEFAULT_DISK_USAGE_LIMIT="85"

# 载入公共函数
source ./check_env_utils.sh

# 检查docker状态和版本
function check_docker_version() {
    docker_install_message_code=""
    docker_service_status=${STATUS_SERVER_NOT_INSTALL}
    docker_version=""
    docker_version_command="docker --version"
    docker_status_command="systemctl status docker"
    # 检查docker是否安装
    docker_status_str=$(${docker_status_command} 2>/dev/null | grep -E "Active:" \
                                                             | awk -F ':' '{print $2}' \
                                                             | awk -F 'since' '{print $1}' \
                                                             | sed -e 's/[ ]*$//g' \
                                                             | sed -e 's/^[ ]*//g')
    if [[ "${docker_status_str}" != "" ]]
    then
        if [[ "${docker_status_str}" == "${SERVICE_STATUS_RUNNING}" ]]
        then
            # docker服务正常
            docker_service_status="${STATUS_NORMAL}[$docker_status_str]"
        else
            # docker服务异常
            docker_service_status="${STATUS_ERROR}[$docker_status_str]"
            docker_install_message_code="${ERROR_DOCKER_SERVICE_CODE}"
        fi

        docker_version=$(${docker_version_command} 2>/dev/null | awk -F '[" ",]' '{print $3}')
        # 找不到docker命令
        if [[ "${docker_version}" == "" ]]
        then
            docker_service_status="${STATUS_ERROR}"
            docker_install_message_code="${ERROR_DOCKER_COMMAND_CODE}"
        else
            if [[ "${docker_service_status}" =~ ${STATUS_NORMAL} ]]
            then
                docker_install_message_code="${INFO_DOCKER_VERSION_RECOMMEND_CODE}"
            fi
        fi
    else
        docker_install_message_code="${ERROR_DOCKER_NOT_INSTALL_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "docker" "${docker_service_status}" "${docker_version}" "${docker_install_message_code}"
}

function check_docker_runtime() {
    runtime_message_code=""
    docker_runtime="$(echo "${docker_info}" | grep 'Runtimes' | awk -F ':' '{print $2}' | sed 's/^ //g')"
    if [[ ! "${docker_runtime}" =~ ${ascend_runtime} ]]
    then
        runtime_message_code="${WARNING_NOT_ASCEND_DOCKER_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "docker runtime" "${docker_runtime}" "" "${runtime_message_code}"
}

function check_image_disk_usage() {
    image_disk_message_code=" "
    if [[ "" == "${docker_info}" ]]
    then
        image_disk_message_code="${ERROR_DOCKER_SERVICE_CODE}"
    else
        docker_root_dir="$(echo "${docker_info}" | grep "Docker Root Dir" | awk -F ':' '{print $2}' | sed 's/^ //g')"
        if [[ "" == "${docker_root_dir}" ]]
        then
            image_disk_message_code="${ERROR_DOCKER_SERVICE_CODE}"
        else
            disk_usage_content="$(df -h "${docker_root_dir}" | awk '{print $1"|"$2"|"$3"|"$4"|"$5"|"$6}' | column -s '|' -t)"
            usage="$(echo "${disk_usage_content}" | grep '/dev' | awk -F ' ' '{print $5}' | awk -F '%' '{print $1}')"
            disk_usage_content="$(echo "${disk_usage_content}" | sed 's/^ //g' | sed 's/^/ *|*/g' | sed 's/$/*|* *|* *|/g')"
            if (( "${DEFAULT_DISK_USAGE_LIMIT}" <= "${usage}" ))
            then
                image_disk_message_code="${WARNING_IMAGE_DISK_USAGE_CODE}"
            fi
            local content="disk usage of docker images*|* *|* *|*${image_disk_message_code}*|\n${disk_usage_content}"
            write_multiline_to_file "${tmp_output_file}" "${content}"
            return
        fi
    fi

    write_single_line_to_file "${tmp_output_file}" "disk usage of docker images" "" "" "${image_disk_message_code}"
}

function do_check() {
    check_docker_version
    # 非master节点执行docker runtime检查
    if [[ "${MASTER_NODE}" != "${nodeType}" ]]
    then
        check_docker_runtime
    fi
    check_image_disk_usage
}

do_check