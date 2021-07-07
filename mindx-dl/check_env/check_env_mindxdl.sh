#!/usr/bin/env bash

# 输入参数
nodeType="$1"
hardWare="$2"
# host_ip="$3"
tmp_output_file="$4"

# 消息码
# error类
ERROR_VOLCANO_ADMISSION_INIT_SERVICE_CODE="error_5001"
ERROR_POD_STATUS_CODE="error_5002"
ERROR_NO_DIR_CODE="error_5003"
ERROR_MISSING_LABEL_CODE="error_5004"
ERROR_NO_IMAGE_CODE="error_5005"
ERROR_SERVICE_NOT_INSTALL_CODE="error_5006"
# 与检查docker中的消息码保持一致
ERROR_DOCKER_SERVICE_CODE="error_3002"
# warning类

# info类
INFO_SERVICE_LOG_DIR_CODE="info_5001"
INFO_NO_NPU_FOUND_CODE="info_5002"
# 与检查k8s中的消息码一致
INFO_PERMISSION_DENIED_CODE="info_4002"

# 常量
# MindX DL服务名称
VOLCANO_SERVICE_NAME="Volcano"
DEVICE_PLUGIN_SERVICE_NAME="Ascend-Device-Plugin"
CADVISOR_SERVICE_NAME="cAdvisor"
HCCL_SERVICE_NAME="HCCL-Controller"
VOLCANO_SCHEDULER_SERVICE="volcano-scheduler"
VOLCANO_CONTROLLERS_SERVICE="volcano-controllers"
VOLCANO_ADMISSION_SERVICE="volcano-admission"
VOLCANO_ADMISSION_INIT_SERVICE="volcano-admission-init"
# 日志目录
HCCL_LOG_DIR="/var/log/mindx-dl/hccl-controller"
VOLCANO_LOG_DIR="/var/log/mindx-dl/volcano*"
DEVICEPLUGIN_LOG_DIR="/var/log/devicePlugin"
CADVISOR_LOG_DIR="/var/log/cadvisor"

# 节点标签
host_arch=$(arch)
if [[ "${host_arch}" == "x86_64" ]]
then
    HOST_ARCH_LABEL="host-arch=huawei-x86"
else
    HOST_ARCH_LABEL="host-arch=huawei-arm"
fi

# node label
# A910
A910_LABEL="accelerator=huawei-Ascend910"
# A310
A310_LABEL="accelerator=huawei-Ascend310"
# A710
A710_LABEL="accelerator=huawei-Ascend710"
# mindx dl master
MINDXDL_MASTER_LABEL="masterselector=dls-master-node"
# mindx dl worker
MINDXDL_WORKER_LABEL="workerselector=dls-worker-node"
# k8s worker
K8S_WORKER_LABEL="node-role.kubernetes.io/worker=worker"
# 300T
A300T_LABEL="accelerator-type=card"


# 载入公共函数
source ./check_env_utils.sh

function check_service_by_name() {
    # pod通配符
    service_name_regexp="$1"
    # 服务使用镜像的名称
    image_name="$2"
    # 服务名
    service_name="$3"
    # all pods信息列表
    all_pods_info_str="$4"
    # 服务状态
    service_status=${STATUS_SERVER_NOT_INSTALL}
    # 服务检查消息码
    service_message_code=""
    # 镜像版本
    service_image=""

    # 有权限访问kubernetes服务
    if [[ "" == "$(echo "${all_pods_info_str}" | grep "was refused")" ]]
    then
        service_status_str=$(echo "${all_pods_info_str}" | grep "$(hostname)" \
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
                    service_message_code="${ERROR_VOLCANO_ADMISSION_INIT_SERVICE_CODE}"
                else
                    status=${STATUS_NORMAL}
                fi
            else
                if [[ "${pod_status}" != "${POD_STATUS_RUNNING}" ]]
                then
                    status=${STATUS_ERROR}
                    service_message_code="${ERROR_POD_STATUS_CODE}"
                else
                    status=${STATUS_NORMAL}
                fi
            fi
            service_status="${status}[$pod_status]"
            write_single_line_to_file "${tmp_output_file}" "${service_name}" "${service_status}" "${service_image}" "${service_message_code}"
        else
            docker_image_str=$(${DOCKER_IMAGES} 2>&1)
            error_str="$(echo "${docker_image_str}" | grep -E "Cannot connect to the Docker daemon|not install")"
            if [[ "${error_str}" != "" ]]
            then
                # docker没启动
                service_message_code="${ERROR_DOCKER_SERVICE_CODE}"
            else
                service_image=$(echo "${docker_image_str}" 2>/dev/null | grep -E "${image_name} " \
                                                                       | awk -F ' ' '{print $1":"$2}')
                service_message_code="${ERROR_SERVICE_NOT_INSTALL_CODE}"
                if [[ "${service_image}" == "" ]]
                then
                  service_image="${NO_IMAGE}"
                  service_message_code="${ERROR_NO_IMAGE_CODE}"
                  write_single_line_to_file "${tmp_output_file}" "${service_name}" "${service_status}" "${service_image}" "${service_message_code}"
                  return
                fi
                service_image="$(echo "${service_image}" | sed 's/^/ *|* *|* *|*/g' | sed 's/$/*|* *|/g')"
            fi
            write_multiline_to_file "${tmp_output_file}" " *|*${service_name}*|*${service_status} *|* *|*${service_message_code} *|\n${service_image}"
        fi
    else
        # 没有权限访问k8s
        service_status=${STATUS_PERMISSION_DENIED}
        service_message_code="${INFO_PERMISSION_DENIED_CODE}"
        write_single_line_to_file "${tmp_output_file}" "${service_name}" "${service_status}" "${service_image}" "${service_message_code}"
    fi
}

# 检查device-plugin状态，版本
function check_device_plugin_service() {
    # pod名称通配符
    service_name_regexp="ascend-device-plugin(.*)-daemonset-"
    # 服务使用镜像的名称
    image_name="ascend-k8sdeviceplugin"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${DEVICE_PLUGIN_SERVICE_NAME}" "$all_pods_info_str"
}

# 检查cadvisor状态，版本
function check_cadvisor_service() {
    # pod名称通配符
    service_name_regexp="cadvisor-"
    # 服务使用镜像的名称
    image_name="google/cadvisor"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${CADVISOR_SERVICE_NAME}" "$all_pods_info_str"
}

# 检查hccl状态，版本
function check_hccl_service() {
    # pod名称通配符
    service_name_regexp="hccl-controller-"
    # 服务使用镜像的名称
    image_name="hccl-controller"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${HCCL_SERVICE_NAME}" "$all_pods_info_str"
}

# 检查volcano-admission状态，版本
function check_volcano_admission_service() {
    # volcano_admission
    service_name=${VOLCANO_ADMISSION_SERVICE}
    # pod名称通配符
    service_name_regexp="volcano-admission-(.*)"
    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-webhook-manager"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${service_name}" "$all_pods_info_str"
}

# 检查volcano-controllers-状态，版本
function check_volcano_controllers_service() {
    # volcano-controllers
    service_name=${VOLCANO_CONTROLLERS_SERVICE}
    # pod名称通配符
    service_name_regexp="volcano-controllers-(.*)"
    # volcano-controllers服务使用镜像的名称
    image_name="volcanosh/vc-controller-manager"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${service_name}" "$all_pods_info_str"
}

# 检查volcano-scheduler状态，版本
function check_volcano_scheduler_service() {
    # volcano-scheduler
    service_name=${VOLCANO_SCHEDULER_SERVICE}
    # pod名称通配符
    service_name_regexp="volcano-scheduler-(.*)"
    # volcano_scheduler服务使用镜像的名称
    image_name="volcanosh/vc-scheduler"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${service_name}" "$all_pods_info_str"
}

# 检查volcano-admission-init状态，版本
function check_volcano_admission_init_service() {
    # volcano-admission-init
    service_name=${VOLCANO_ADMISSION_INIT_SERVICE}
    # pod名称通配符
    service_name_regexp="volcano-admission-init-(.*)"
    # volcano_admission服务使用镜像的名称
    image_name="volcanosh/vc-webhook-manager"
    # all pods信息列表
    all_pods_info_str="$1"

    # 检查
    check_service_by_name "${service_name_regexp}" "${image_name}" "${service_name}" "$all_pods_info_str"
}

# 检查MindX DL组件版本和状态
function check_mindxdl_version() {
    # all pods信息列表，避免多次获取影响速度
    all_pods_info_str="$(${UNSET_PROXY} && ${GET_ALL_PODS_COMMAND} 2>&1)"

    # master节点
    if [[ "${nodeType}" =~ ${MASTER_NODE}(.*) ]]
    then
        check_hccl_service "${all_pods_info_str}"
        check_volcano_admission_service "${all_pods_info_str}"
        check_volcano_admission_init_service "${all_pods_info_str}"
        check_volcano_controllers_service "${all_pods_info_str}"
        check_volcano_scheduler_service "${all_pods_info_str}"
    fi
    # worker节点
    if [[ "${nodeType}" =~ (.*)${WORKER_NODE} ]]
    then
        check_device_plugin_service "${all_pods_info_str}"
        check_cadvisor_service "${all_pods_info_str}"
    fi
}

function check_service_log_dir() {
    log_dir="$1"
    service_name="$2"

    message_code=""
    log_dir_str="$(ls -dl ${log_dir} 2>/dev/null)"
    local content=""
    if [[ "" == "${log_dir_str}" ]]
    then
        message_code="${ERROR_NO_DIR_CODE}"
        content="$(echo "${service_name}:" | sed 's/^/ *|* *|*/g' | sed "s/$/*|* *|*${message_code}*|/g")"
    else
        content="$(echo "${log_dir_str}" | sed "1i ${service_name}:" | sed 's/^/ *|* *|*/g' | sed 's/$/*|* *|* *|/g')"
    fi

    write_multiline_to_file "${tmp_output_file}" "${content}"
}

# 检查MindX dl日志目录
function check_mindxdl_log_dir() {
    write_single_line_to_file "${tmp_output_file}" "service log dir" "" "" "${INFO_SERVICE_LOG_DIR_CODE}"
    write_single_line_to_file "${tmp_output_file}"

    # master节点
    if [[ "${nodeType}" =~ ${MASTER_NODE}(.*) ]]
    then
        # hccl
        check_service_log_dir "${HCCL_LOG_DIR}" "${HCCL_SERVICE_NAME}"
        write_single_line_to_file "${tmp_output_file}"
        # volcano
        check_service_log_dir "${VOLCANO_LOG_DIR}" "${VOLCANO_SERVICE_NAME}"
        write_single_line_to_file "${tmp_output_file}"
    fi
    # worker节点
    if [[ "${nodeType}" =~ (.*)${WORKER_NODE} ]]
    then
        # device plugin
        check_service_log_dir "${DEVICEPLUGIN_LOG_DIR}" "${DEVICE_PLUGIN_SERVICE_NAME}"
        write_single_line_to_file "${tmp_output_file}"
        #cadvisor
        check_service_log_dir "${CADVISOR_LOG_DIR}" "${CADVISOR_SERVICE_NAME}"
    fi

}

# 检查由device plugin发现的、本机可用的芯片数量
function check_nup_available_by_service() {
    allocatable_message_code=""
    allocatable_npu_num="0"
    used_npu_num="0"
    discovered_npu_status=""
    node_describe_info="$(${UNSET_PROXY} && kubectl describe node "$(hostname)" 2>&1)"
    if [[ "" != "$(echo "${node_describe_info}" | grep "was refused")" ]]
    then
        allocatable_message_code="${INFO_PERMISSION_DENIED_CODE}"
        discovered_npu_status="${STATUS_PERMISSION_DENIED}"
    else
        if [[ "${hardWare}" == "${HW_300T}" ]] || [[ "${hardWare}" == "${HW_TRAIN}" ]]
        then
            resource_type="huawei.com/Ascend910"
        elif [[ "${hardWare}" == "${HW_300I_PRO}" ]]
        then
            resource_type="huawei.com/Ascend710"
        else
            resource_type="huawei.com/Ascend310"
        fi
        # 分配给整个节点的、可用的总数
        allocatable_npu_num="$(echo "${node_describe_info}" | grep -B 20 "System Info:" \
                                                            | grep -A 20 "Allocatable:" \
                                                            | grep "${resource_type}" \
                                                            | awk -F ":" '{print $2}' \
                                                            | sed 's/ //g')"
        used_npu_num="$(echo "${node_describe_info}" | grep -B 20 "Events" \
                                                     | grep -A 20 "Allocated resources:" \
                                                     | grep "${resource_type}" \
                                                     | awk '{print $2}' \
                                                     | sed 's/ //g')"

        if [[ "" == "${allocatable_npu_num}" ]] || [[ "0" == "${allocatable_npu_num}" ]]
        then
            allocatable_npu_num="0"
            allocatable_message_code="${INFO_NO_NPU_FOUND_CODE}"
        fi

        if [[ "" == "${used_npu_num}" ]]
        then
            used_npu_num="0"
        fi
        discovered_npu_status="total: ${allocatable_npu_num}, used: ${used_npu_num}"
    fi

    write_single_line_to_file "${tmp_output_file}" "number of npu in node" "${discovered_npu_status}" "" "${allocatable_message_code}"
}

function check_node_label() {
    worker_label_filter=" "
    master_label_filter=" "
    local label_message_code=""

    # master节点
    if [[ "${nodeType}" =~ ${MASTER_NODE}(.*) ]]
    then
        master_label_filter="${MINDXDL_MASTER_LABEL}"
    fi

    # worker节点
    if [[ "${nodeType}" =~ (.*)${WORKER_NODE} ]]
    then
        worker_base_label="${MINDXDL_WORKER_LABEL}|${K8S_WORKER_LABEL}|${HOST_ARCH_LABEL}"
        case "${hardWare}" in
            "${HW_300T}")
                # 300T标签
                worker_label_filter="${worker_base_label}|${A910_LABEL}|${A300T_LABEL}"
                ;;
            "${HW_TRAIN}")
                # 910训练标签
                worker_label_filter="${worker_base_label}|${A910_LABEL}"
                ;;
            "${HW_INFER}")
                # 310推理标签
                worker_label_filter="${worker_base_label}|${A310_LABEL}"
                ;;
            "${HW_300I_PRO}")
                # A300I Pro标签
                worker_label_filter="${worker_base_label}|${A710_LABEL}"
                ;;
        esac
    fi

    label_filter="${master_label_filter}|${worker_label_filter}"
    # 根据节点类型和硬件形态得到的标准的标签，表现为行的形式，
    label_filter_rows="$(echo "${label_filter}" | sed "s/|/\\n/g" | grep -v " ")"

    filtered_label=""
    node_labels_str="$(${UNSET_PROXY} && kubectl get nodes "$(hostname)" --show-labels=true 2>&1)"
    if [[ "" != "$(echo "${node_labels_str}" | grep "was refused")" ]]
    then
        label_message_code="${INFO_PERMISSION_DENIED_CODE}"
        label_status="${STATUS_PERMISSION_DENIED}"
        write_single_line_to_file "${tmp_output_file}" "node label" "${label_status}" "" "${label_message_code}"
    else
        # 实际查出来的标签
        filtered_label="$(echo "${node_labels_str}" | awk '{print $6}' | sed "s/,/\\n/g" | grep -E "${label_filter}")"
        # 缺失的标签
        missed_label="$(echo -e "${filtered_label}\n${label_filter_rows}" | grep -vE "^$" | sort | uniq -u)"

        if [[ "" == "${missed_label}" ]]
        then
            label_status="label meets the requirements"
            write_single_line_to_file "${tmp_output_file}" "node label" "${label_status}"
        else
            label_message_code="${ERROR_MISSING_LABEL_CODE}"
            label_status="$(echo -e "missing label:\n${missed_label}" | sed 's/^/ *|* *|*/g' | sed 's/$/*|* *|* *|/g')"
            local content=" *|*node label*|* *|* *|*${label_message_code}*|\n${label_status}"
            write_multiline_to_file "${tmp_output_file}" "${content}"
        fi
    fi

}

function do_check() {
    write_category "mindxdl-related"
    check_mindxdl_version
    check_mindxdl_log_dir
    # worker节点检查device-plugin识别的npu
    if [[ "${nodeType}" =~ (.*)${WORKER_NODE} ]]
    then
        check_nup_available_by_service
    fi
    check_node_label
}

do_check