#!/usr/bin/env bash
# Copyright © Huawei Technologies Co., Ltd. 2020. All rights reserved.

# 全局参数
# 输出文件
file_dir=$(dirname "$(readlink -f "$0")")
file_path=${file_dir}"/env_check_report.txt"

SERVICE_CHECK_STR_ARRAY=()
# MindX DL服务名称
DEVICE_PLUGIN_SERVICE_NAME="Ascend-Device-Plugin"
CADVISOR_SERVICE_NAME="cAdvisor"
HCCL_SERVICE_NAME="HCCL-Controller"
VOLCANO_SCHEDULER_SERVICE="volcano-scheduler"
VOLCANO_CONTROLLERS_SERVICE="volcano-controllers"
VOLCANO_ADMISSION_SERVICE="volcano-admission"
VOLCANO_ADMISSION_INIT_SERVICE="volcano-admission-init"

# 格式化输出参数,列宽
service_col_length=7
status_col_length=6
version_col_length=7
default_ip_col_width=2

# 信息常量
STATUS_NORMAL="Normal"
STATUS_ERROR="Error"
STATUS_SERVER_NOT_INSTALL="Not install"
STATUS_POD_COMPLETED="Completed"
STATUS_PERMISSION_DENIED="Can't get service status, permission denied"
POD_STATUS_RUNNING="Running"
SERVICE_STATUS_RUNNING="active (running)"
NO_IMAGE="No image"
DOCKER_SERVICE_NOT_RUNNING="Docker service not running"

# 操作命令
UNSET_PROXY="unset http_proxy https_proxy"
GET_ALL_PODS_COMMAND="kubectl get pods --all-namespaces -o wide"
DESCRIBE_POD_COMMAND="kubectl describe pod"
NPU_INFO_COMMAND="npu-smi info"
DOCKER_IMAGES="docker images"

# 节点类型常量
MASTER_NODE="master"
WORKER_NODE="worker"
# 既是master又是worker
MASTER_WORKER_NODE="master-worker"

# 检查输入参数合法性
function check_input_params() {
    node_type=$1
    ip_str=$2

    if [[ ${node_type} != "${MASTER_NODE}" ]] &&\
        [[ ${node_type} != "${WORKER_NODE}" ]] &&\
        [[ ${node_type} != "${MASTER_WORKER_NODE}" ]]
    then
        echo -e "\nError: The first parameter is node_type. It can only be \"master\", \"worker\", or \"master-worker\".\n"
        exit 1
    fi

    ip_regexp="^([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])$"
    if [[ ! ${ip_str} =~ ${ip_regexp} ]] &&\
        [[ ${ip_str} !=  "" ]]
    then
        echo -e "\nError: The second parameter is ip address. Enter a correct IP address or leave it empty.\n"
        exit 1
    fi
}
# 入参检查
check_input_params "$1" "$2"
# 节点类型
check_node_type=$1
# 传入的ip
ip_addr=$2

function set_max_len() {
    service_name_len=$(echo "$1" | awk -F '|' '{print length($1)}')
    status_len=$(echo "$1" | awk -F '|' '{print length($2)}')
    version_len=$(echo "$1" | awk -F '|' '{print length($3)}')

    if (( service_name_len > service_col_length ))
    then
        service_col_length=${service_name_len}
    fi

    if (( status_len > status_col_length ))
    then
        status_col_length=${status_len}
    fi

    if (( version_len > version_col_length ))
    then
        version_col_length=${version_len}
    fi
}

# 检查NPU驱动状态和版本
function check_npu() {
    service_name="npu-driver"
    npu_driver_status=${STATUS_SERVER_NOT_INSTALL}
    npu_driver_version=""
    NPU_INFO_COMMAND="npu-smi info"
    ${NPU_INFO_COMMAND} > /dev/null 2>&1

    if ${NPU_INFO_COMMAND}
    then
        npu_driver_version=$(${NPU_INFO_COMMAND} 2>/dev/null | grep -E 'npu-smi(.*)Version:' \
                                                             | awk -F ':' '{print $2}' \
                                                             | awk -F ' ' '{print $1}' \
                                                             | sed -e 's/[ ]*$//g' \
                                                             | sed -e 's/^[ ]*//g')
        npu_driver_status=${STATUS_NORMAL}
    fi

    check_result_str="${service_name}|${npu_driver_status}|${npu_driver_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检查docker状态和版本
function check_docker() {
    service_name="docker"
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
            status=${STATUS_NORMAL}
        else
            # docker服务异常
            status=${STATUS_ERROR}
        fi
        docker_service_status="${status}[$docker_status_str]"

        docker_version=$(${docker_version_command} 2>/dev/null | awk -F '[" ",]' '{print $3}')
        # 找不到docker命令
        if [[ "${docker_version}" = "" ]]
        then
            docker_version="${STATUS_ERROR}[The Docker command cannot be found. The Docker environment may be damaged.]"
        fi
    fi

    check_result_str="${service_name}|${docker_service_status}|${docker_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检测kubelet状态，版本
function check_kubelet_service() {
    service_name="kubelet"
    kubelet_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubelet_version=""
    kubelet_version_command="kubelet --version"
    kubelet_status_command="systemctl status kubelet"

    kubelet_status_str=$(${kubelet_status_command} 2>/dev/null | grep -E "Active:" \
                                                               | awk -F ':' '{print $2}' \
                                                               | awk -F 'since' '{print $1}' \
                                                               | sed 's/[ ]*$//g' \
                                                               | sed 's/^[ ]*//g')

    if [ "${kubelet_status_str}" != "" ]
    then
        if [[ "${kubelet_status_str}" == "${SERVICE_STATUS_RUNNING}" ]]
        then
            # kubelet服务正常
            status=${STATUS_NORMAL}
        else
            # kubelet服务异常
            status=${STATUS_ERROR}
        fi
        kubelet_service_status="${status}[$kubelet_status_str]"

        kubelet_version=$(${kubelet_version_command} 2>/dev/null | awk -F '[" ",]' '{print $2}')
        # 找不到kubelet命令
        if [[ "${kubelet_version}" = "" ]]
        then
            kubelet_version="${STATUS_ERROR}[The kubelet command cannot be found. The kubelet environment may be damaged.]"
        fi
    fi

    check_result_str="${service_name}|${kubelet_service_status}|${kubelet_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检测kubeadm状态，版本
function check_kubeadm_service() {
    service_name="kubeadm"
    kubeadm_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubeadm_version=""
    kubeadm_version_command="kubeadm version"

    kubeadm_version=$(${kubeadm_version_command} 2>/dev/null | awk -F '[,:]' '{print $7}' \
                                                             | sed -e 's/"//g')
    # 存在kubeadm命令
    if [[ "${kubeadm_version}" != "" ]]
    then
        kubeadm_service_status=${STATUS_NORMAL}
        kubeadm_version="${kubeadm_version}"
    fi

    check_result_str="${service_name}|${kubeadm_service_status}|${kubeadm_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检测kubectl状态，版本
function check_kubectl_service() {
    service_name="kubectl"
    kubectl_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubectl_version=""
    kubectl_version_command="kubectl version"

    kubectl_version=$(${kubectl_version_command} 2>/dev/null | head -n 1 \
                                                             | awk -F '[,:]' '{print $7}' \
                                                             | sed -e 's/"//g')
    # 存在kubectl命令
    if [[ "${kubectl_version}" != "" ]]
    then
        kubectl_service_status=${STATUS_NORMAL}
        kubectl_version=${kubectl_version}
    fi

    check_result_str="${service_name}|${kubectl_service_status}|${kubectl_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

function check_k8s() {
    check_kubelet_service
    check_kubeadm_service
    check_kubectl_service
}

function check_mindxdl_by_name() {
    # pod通配符
    service_name_regexp=$1
    # 服务使用镜像的名称
    image_name=$2
    # 服务名
    service_name=$3
    # 服务状态
    service_status=${STATUS_SERVER_NOT_INSTALL}
    pods_status_str=$(${UNSET_PROXY} && ${GET_ALL_PODS_COMMAND} 2>&1)

    # 有权限访问kubernetes服务
    if [[ ! "${pods_status_str}" =~ (.*was refused.*) ]]
    then
        service_status_str=$(echo "${pods_status_str}" | grep "$(hostname)" \
                                                       | grep -E "${service_name_regexp}" \
                                                       | head -n 1 \
                                                       | sed -e 's/[ ]*$//g' \
                                                       | sed -e 's/^[ ]*//g')
        # 安装了对应的服务
        if [[ ${service_status_str} != "" ]]
        then
            # 使用的k8s命名空间
            pod_namespace=$(echo "${service_status_str}" | awk -F ' ' '{print $1}')
            # pod名称
            pod_name=$(echo "${service_status_str}" | awk -F ' ' '{print $2}')
            # pod状态
            pod_status=$(echo "${service_status_str}" | awk -F ' ' '{print $4}')
            # 查询pod详情
            describe_pod_command="${DESCRIBE_POD_COMMAND} -n ${pod_namespace} ${pod_name}"
            # pod使用的镜像
            service_image=$(${UNSET_PROXY} && ${describe_pod_command} 2>/dev/null | grep "Image:" \
                                                                                  | awk -F ' ' '{print $2}' \
                                                                                  | sed -e 's/[ ]*$//g' \
                                                                                  | sed -e 's/^[ ]*//g')
            if [[ ${service_name} == "${VOLCANO_ADMISSION_INIT_SERVICE}" ]]
            then
                if [[ "${pod_status}" != "${STATUS_POD_COMPLETED}" ]]
                then
                    status=${STATUS_ERROR}
                else
                    status=${STATUS_NORMAL}
                fi
            else
                if [[ "${pod_status}" != "${POD_STATUS_RUNNING}" ]]
                then
                    status=${STATUS_ERROR}
                else
                    status=${STATUS_NORMAL}
                fi
            fi
            service_status="${status}[$pod_status]"
        else
            docker_image_str=$(${DOCKER_IMAGES} 2>&1)
            if [[ "${docker_image_str}" =~ (Cannot connect to the Docker daemon.*) ]] \
                || [[ "${docker_image_str}" =~ (.*not install) ]]
            then
                # docker没启动
                service_image=${DOCKER_SERVICE_NOT_RUNNING}
            else
                service_image=$(echo "${docker_image_str}" 2>/dev/null | grep -E "${image_name} " \
                                                                       | awk -F ' ' '{print $1":"$2}' \
                                                                       | xargs \
                                                                       | sed -e 's/ /,/g')
              if [[ "${service_image}" == "" ]]
              then
                  service_image=${NO_IMAGE}
              fi
            fi
        fi
    else
        # 没有权限访问k8s
        service_status=${STATUS_PERMISSION_DENIED}
        service_image=""
    fi

    check_result_str="${service_name}|${service_status}|${service_image}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检查device-plugin状态，版本
function check_device_plugin_service() {
    # pod名称通配符
    service_name_regexp="ascend-device-plugin(.*)-daemonset-"

    # 服务使用镜像的名称
    image_name="ascend-k8sdeviceplugin"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${DEVICE_PLUGIN_SERVICE_NAME}"
}

# 检查cadvisor状态，版本
function check_cadvisor_service() {
    # pod名称通配符
    service_name_regexp="cadvisor-"

    # 服务使用镜像的名称
    image_name="google/cadvisor"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${CADVISOR_SERVICE_NAME}"
}

# 检查hccl状态，版本
function check_hccl_service() {
    # pod名称通配符
    service_name_regexp="hccl-controller-"

    # 服务使用镜像的名称
    image_name="hccl-controller"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${HCCL_SERVICE_NAME}"
}

# 检查volcano-admission状态，版本
function check_volcano_admission_service() {
    # volcano_admission
    service_name=${VOLCANO_ADMISSION_SERVICE}

    # pod名称通配符
    service_name_regexp="volcano-admission-(.*)"

    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-webhook-manager"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${service_name}"
}

# 检查volcano-controllers-状态，版本
function check_volcano_controllers_service() {
    # volcano-controllers
    service_name=${VOLCANO_CONTROLLERS_SERVICE}

    # pod名称通配符
    service_name_regexp="volcano-controllers-(.*)"

    # volcano-controllers服务使用镜像的名称
    image_name="volcanosh/vc-controller-manager"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${service_name}"
}

# 检查volcano-scheduler状态，版本
function check_volcano_scheduler_service() {
    # volcano-controllers
    service_name=${VOLCANO_SCHEDULER_SERVICE}

    # pod名称通配符
    service_name_regexp="volcano-scheduler-(.*)"

    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-scheduler"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${service_name}"
}

# 检查volcano-admission-init状态，版本
function check_volcano_admission_init_service() {
    # volcano-admission-init
    service_name=${VOLCANO_ADMISSION_INIT_SERVICE}

    # pod名称通配符
    service_name_regexp="volcano-admission-init-(.*)"

    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-webhook-manager"

    # 检查
    check_mindxdl_by_name "${service_name_regexp}" "${image_name}" "${service_name}"
}

# 检查MindX DL组件状态和版本
function check_mindxdl() {
    # master节点
    if [[ "${check_node_type}" =~ ${MASTER_NODE}(.*) ]]
    then
        check_hccl_service
        check_volcano_admission_service
        check_volcano_admission_init_service
        check_volcano_controllers_service
        check_volcano_scheduler_service
    fi
    # worker节点
    if [[ "${check_node_type}" =~ (.*)${WORKER_NODE} ]]
    then
        check_device_plugin_service
        check_cadvisor_service
    fi

}

function print_format_to_file() {
    rm -rf "${file_path}"

    # 检查结果数组长度
    arr_length=${#SERVICE_CHECK_STR_ARRAY[@]}
    # 打印时设置的其他符号长度
    other_symbel_length=16

    hostname=$(hostname)
    # hostname长度
    os_hostname_col_length=$(echo "${hostname}" | awk '{print length($0)}')
    # “hostname”字符长度
    table_head_hostname_col_length=8
    # hostname列宽
    if (( os_hostname_col_length > table_head_hostname_col_length ))
    then
        hostname_col_length=${os_hostname_col_length}
    else
        hostname_col_length=${table_head_hostname_col_length}
    fi

    ip_addr_length=$(echo "${ip_addr}" | awk '{print length($0)}')
    ip_col_length=${default_ip_col_width}
    if (( ip_addr_length > default_ip_col_width ))
    then
        ip_col_length=${ip_addr_length}
    fi

    row_max_length=$((service_col_length + ip_col_length + \
                        status_col_length + version_col_length + \
                        hostname_col_length + other_symbel_length))

    printf "%-${row_max_length}s\n" "-" | sed -e 's/ /-/g' >> "${file_path}"
    for((i=0; i<arr_length; i++));
    do
        if (( i == 0 ))
        then
            table_hostname=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $1}')
            table_ip_addr=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $2}')
            table_service_name=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $3}')
            table_status=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $4}')
            table_version=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $5}')
        else
            table_hostname=${hostname}
            table_ip_addr=${ip_addr}
            table_service_name=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $1}')
            table_status=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $2}')
            table_version=$(echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $3}')
        fi

        printf "| %-${hostname_col_length}s | %-${ip_col_length}s | %-${service_col_length}s | %-${status_col_length}s | %-${version_col_length}s |\n" \
                "${table_hostname}" "${table_ip_addr}" "${table_service_name}" "${table_status}" "${table_version}" >> "${file_path}"
        printf "%-${row_max_length}s\n" "-" | sed -e 's/ /-/g' >> "${file_path}"
    done
}

# 执行检查
function do_check() {
    # 检查npu
    check_npu
    # 检查docker
    check_docker
    # 检查k8s
    check_k8s
    # 检查MindX DL组件
    check_mindxdl
}

function main() {
    # 表头
    table_head="hostname|ip|service|status|version"
    set_max_len ${table_head}
    # 添加表头
    SERVICE_CHECK_STR_ARRAY+=("${table_head}")
    # 检查
    do_check
    # 格式化打印
    print_format_to_file

    chmod 540 "${file_path}"
}

main >>/dev/null 2>&1
cat "${file_path}"
echo ""
echo "Finished! The check report is stored in the ${file_path}"
echo ""