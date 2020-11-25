# Copyright © Huawei Technologies Co., Ltd. 2020. All rights reserved.
#!/usr/bin/env bash

# 本地检测输出的文件
file_dir=$(dirname $(readlink -f $0))
file_path=${file_dir}"/env_check_report.txt"

SERVICE_CHECK_STR_ARRAY=()
# MindX DL服务名称
DEVICE_PLUGIN_SERVICE_NAME="Ascend-Device-Plugin"
CADVISOR_SERVICE_NAME="cAdvisor"
HCCL_SERVICE_NAME="HCCL-Controller"
VOLCANO_SERVICE_NAME="Volcano"

# 格式化输出参数,列宽
service_col_length=0
status_col_length=0
version_col_length=0
# "ip"字符串长度
default_ip_col_length=2

# 信息
STATUS_NORMAL="normal"
STATUS_ERROR="error"
SERVER_NOT_INSTALL="not install"
POD_STATUS_RUNNING="Running"
SERVICE_STATUS_RUNNING="active (running)"
NO_IMAGE="no image"

# 操作命令
GET_ALL_PODS_COMMAND="unset http_proxy https_proxy && kubectl get pods -A -o wide"
DESCRIBE_POD_COMMAND="unset http_proxy https_proxy && kubectl describe pod"
NPU_INFO_COMMAND="npu-smi info"
DOCKER_IMAGES="docker images"
DOKCER_INFO="docker info"

# 检查的节点类型，默认为检查worker
MASTER_NODE="master"
WORKER_NODE="worker"
# 既是master又是worker
MASTER_WORKER_NODE="master-worker"
if [[ $1 == "${MASTER_NODE}" ]]
then
    check_node_type=${MASTER_NODE}
elif [[ $1 == "${MASTER_WORKER_NODE}" ]]
then
    check_node_type=${MASTER_WORKER_NODE}
else
    check_node_type=${WORKER_NODE}
fi

# 是否传入ip参数
if [[ $2 != "" ]]
then
    ip_addr=$2
else
    ip_addr=""
fi

# service,status,version列最大字符串长度
function set_max_len() {
    service_name_len=`echo "$1" | awk -F '|' '{print length($1)}'`
    status_len=`echo "$1" | awk -F '|' '{print length($2)}'`
    version_len=`echo "$1" | awk -F '|' '{print length($3)}'`

    if (( ${service_name_len} > ${service_col_length} ))
    then
        service_col_length=${service_name_len}
    fi

    if (( ${status_len} > ${status_col_length} ))
    then
        status_col_length=${status_len}
    fi

    if (( ${version_len} > ${version_col_length} ))
    then
        version_col_length=${version_len}
    fi
}

# 检查NPU驱动状态和版本
function check_npu() {
    service_name="npu-driver"
    npu_driver_status=${SERVER_NOT_INSTALL}
    npu_driver_version=""
    NPU_INFO_COMMAND="npu-smi info"
    ${NPU_INFO_COMMAND} > /dev/null 2>&1

    if [ $? == 0 ]
    then
        npu_driver_version=`${NPU_INFO_COMMAND} 2>/dev/null | grep -E 'npu-smi(.*)Version:' \
                                                            | awk -F ':' '{print $2}' \
                                                            | awk -F ' ' '{print $1}' \
                                                            | sed -e 's/[ ]*$//g' \
                                                            | sed -e 's/^[ ]*//g'`
        npu_driver_status=${STATUS_NORMAL}
    fi

    check_result_str="${service_name}|${npu_driver_status}|${npu_driver_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检查docker状态和版本
function check_docker() {
    service_name="docker"
    docker_service_status=${SERVER_NOT_INSTALL}
    docker_version=""
    docker_version_command="docker --version"
    docker_status_command="systemctl status docker"
    # 检查docker是否安装
    docker_status_str=`${docker_status_command} 2>/dev/null | grep -E "Active:" \
                                                            | awk -F ':' '{print $2}' \
                                                            | awk -F 'since' '{print $1}' \
                                                            | sed -e 's/[ ]*$//g' \
                                                            | sed -e 's/^[ ]*//g'`
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

        docker_version=`${docker_version_command} 2>/dev/null | awk -F '[" ",]' '{print $3}'`
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
    kubelet_service_status=${SERVER_NOT_INSTALL}
    kubelet_version=""
    kubelet_version_command="kubelet --version"
    kubelet_status_command="systemctl status kubelet"

    kubelet_status_str=`${kubelet_status_command} 2>/dev/null | grep -E "Active:" \
                                                              | awk -F ':' '{print $2}' \
                                                              | awk -F 'since' '{print $1}' \
                                                              | sed 's/[ ]*$//g' \
                                                              | sed 's/^[ ]*//g'`

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

        kubelet_version=`${kubelet_version_command} 2>/dev/null | awk -F '[" ",]' '{print $2}'`
        # 找不到kubelet命令
        if [[ "${kubelet_version}" = "" ]]
        then
            kubelet_version="${STATUS_ERROR}[The kubelet command cannot be found. The kubernetes environment may be damaged.]"
        fi
    fi

    check_result_str="${service_name}|${kubelet_service_status}|${kubelet_version}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检测kubeadm状态，版本
function check_kubeadm_service() {
    service_name="kubeadm"
    kubeadm_service_status=${SERVER_NOT_INSTALL}
    kubeadm_version=""
    kubeadm_version_command="kubeadm version"

    kubeadm_version=`${kubeadm_version_command} 2>/dev/null | awk -F '[,:]' '{print $7}' \
                                                            | sed -e 's/"//g'`
    # 存在到kubeadmr命令
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
    kubectl_service_status=${SERVER_NOT_INSTALL}
    kubectl_version=""
    kubectl_version_command="kubectl version"

    kubectl_version=`${kubectl_version_command} 2>/dev/null | head -n 1 \
                                                            | awk -F '[,:]' '{print $7}' \
                                                            | sed -e 's/"//g'`
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

# 检查K8s状态和版本
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
    service_status=${SERVER_NOT_INSTALL}
    service_status_str=`${GET_ALL_PODS_COMMAND} 2>/dev/null | grep \`hostname\` \
                                                            | grep -E "${service_name_regexp}" \
                                                            | head -n 1 \
                                                            | sed -e 's/[ ]*$//g' \
                                                            | sed -e 's/^[ ]*//g'`
    # 节点部署了服务
    if [[ "${service_status_str}" != "" ]]
    then
        # 使用的k8s命名空间
        pod_namespace=`echo "${service_status_str}" | awk -F ' ' '{print $1}'`
        # pod名称
        pod_name=`echo "${service_status_str}" | awk -F ' ' '{print $2}'`
        # pod状态
        pod_status=`echo "${service_status_str}" | awk -F ' ' '{print $4}'`
        # 查询pod详情
        describe_pod_command="${DESCRIBE_POD_COMMAND} -n ${pod_namespace} ${pod_name}"
        # pod使用的镜像
        service_image=`${describe_pod_command} 2>/dev/null | grep "Image:" \
                                                           | awk -F ' ' '{print $2}' \
                                                           | sed -e 's/[ ]*$//g' \
                                                           | sed -e 's/^[ ]*//g'`
        if [[ "${pod_status}" != "${POD_STATUS_RUNNING}" ]]
        then
            status=${STATUS_ERROR}
        else
            status=${STATUS_NORMAL}
        fi
        service_status="${status}[$pod_status]"
    else
        service_image=`${DOCKER_IMAGES} 2>/dev/null | grep -E "${image_name}" \
                                                    | awk -F ' ' '{print $1":"$2}' \
                                                    | xargs \
                                                    | sed -e 's/ /,/g'`
        if [[ "${service_image}" == "" ]]
        then
            service_image=${NO_IMAGE}
        fi
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
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${DEVICE_PLUGIN_SERVICE_NAME}
}

# 检查cadvisor状态，版本
function check_cadvisor_service() {
    # pod名称通配符
    service_name_regexp="cadvisor-"

    # 服务使用镜像的名称
    image_name="google/cadvisor"

    # 检查
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${CADVISOR_SERVICE_NAME}
}

# 检查hccl状态，版本
function check_hccl_service() {
    # pod名称通配符
    service_name_regexp="hccl-controller-"

    # 服务使用镜像的名称
    image_name="hccl-controller"

    # 检查
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${HCCL_SERVICE_NAME}
}

# 检查volcano-admission状态，版本
function check_volcano_admission_service() {
    # pod名称通配符
    service_name_regexp="volcano-admission-(.*)"

    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-webhook-manager"

    # 检查
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${VOLCANO_SERVICE_NAME}
}

# 检查volcano-controllers-状态，版本
function check_volcano_controllers_service() {
    # pod名称通配符
    service_name_regexp="volcano-controllers-(.*)"

    # volcano-controllers服务使用镜像的名称
    image_name="volcanosh/vc-controller-manager"

    # 检查
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${VOLCANO_SERVICE_NAME}
}

# 检查volcano-scheduler状态，版本
function check_volcano_scheduler_service() {
    # pod名称通配符
    service_name_regexp="volcano-scheduler-(.*)"

    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-scheduler"

    # 检查
    check_mindxdl_by_name ${service_name_regexp} ${image_name} ${VOLCANO_SERVICE_NAME}
}

# 检查volcano-base状态，版本
function check_volcano_base_service() {
    image_name="volcanosh/vc-webhook-manager-base"
    service_image=`${DOCKER_IMAGES} 2>/dev/null | grep -E "${image_name}" \
                                                | awk -F ' ' '{print $1":"$2}' \
                                                | xargs \
                                                | sed -e 's/ /,/g'`

    if [[ "${service_image}" != "" ]]
    then
        service_status=${STATUS_NORMAL}
    else
        service_status=${SERVER_NOT_INSTALL}
        service_image=${NO_IMAGE}
    fi

    check_result_str="${VOLCANO_SERVICE_NAME}|${service_status}|${service_image}"
    set_max_len "${check_result_str}"
    SERVICE_CHECK_STR_ARRAY+=("${check_result_str}")
}

# 检查MindX DL组件状态和版本
function check_mindxdl() {
    # master节点
    if [[ "${check_node_type}" =~ ${MASTER_NODE}(.*) ]]
    then
        check_hccl_service
        check_volcano_admission_service
        check_volcano_base_service
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
    rm -rf ${file_path}

    # 检查结果数组长度
    arr_length=${#SERVICE_CHECK_STR_ARRAY[@]}
    # 打印时设置的其他符号长度
    other_symbel_length=16

    hostname=`hostname`
    # hostname长度
    os_hostname_col_length=`echo ${hostname} | awk '{print length($0)}'`
    # “hostname”字符长度
    table_head_hostname_col_length=8
    # hostname列宽
    if (( ${os_hostname_col_length} > ${table_head_hostname_col_length} ))
    then
        hostname_col_length=${os_hostname_col_length}
    else
        hostname_col_length=${table_head_hostname_col_length}
    fi

    # ip地址长度
    ip_addr_length=$(echo "${ip_addr}" | awk '{print length($0)}')
    # ip列字符串长度
    ip_col_length=${default_ip_col_length}
    if (( ${ip_addr_length} > ${default_ip_col_length} ))
    then
        ip_col_length=${ip_addr_length}
    fi

    # 一行字符串长度
    row_str_length=$((${service_col_length} + ${ip_col_length} + \
                        ${status_col_length} + ${version_col_length} + \
                        ${hostname_col_length} + ${other_symbel_length}))

    printf "%-${row_str_length}s\n" "-" | sed -e 's/ /-/g' >> ${file_path}
    for((i=0; i<${arr_length}; i++));
    do
        if (( $i == 0 ))
        then
            # 表头
            table_hostname=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $1}'`
            table_ip_addr=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $2}'`
            table_service_name=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $3}'`
            table_status=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $4}'`
            table_version=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $5}'`
        else
            table_hostname=${hostname}
            table_ip_addr=${ip_addr}
            table_service_name=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $1}'`
            table_status=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $2}'`
            table_version=`echo "${SERVICE_CHECK_STR_ARRAY[i]}" | awk -F '|' '{print $3}'`
        fi

        printf "| %-${hostname_col_length}s | %-${ip_col_length}s | %-${service_col_length}s | %-${status_col_length}s | %-${version_col_length}s |\n" \
                "${table_hostname}" "${table_ip_addr}" "${table_service_name}" "${table_status}" "${table_version}" >> ${file_path}
        printf "%-${row_str_length}s\n" "-" | sed -e 's/ /-/g' >> ${file_path}
    done
}

#执行检查
function do_check() {
    # 表头
    table_head="hostname|ip|service|status|version"
    set_max_len ${table_head}
    # 添加表头
    SERVICE_CHECK_STR_ARRAY+=(${table_head})

    # 检查npu
    check_npu
    # 检查docker
    check_docker
    # 检查k8s
    check_k8s

    docker_info_str=`${DOKCER_INFO}`
    # docker没启动就不检查MindX DL组件
    if [[ "${docker_info_str}" =~ (.*Cannot connect to the Docker daemon) ]] \
        || [[ "${docker_status}" =~(.*not install) ]]
    then
        return
    fi
    # 检查MindX DL组件
    check_mindxdl
}

function main() {
    # 检查
    do_check
    # 格式化打印
    print_format_to_file

    chmod 540 ${file_path}
}

main >>/dev/null 2>&1
cat ${file_path}
echo ""
echo "Finished! The check report is stored in the ${file_path}"
echo ""