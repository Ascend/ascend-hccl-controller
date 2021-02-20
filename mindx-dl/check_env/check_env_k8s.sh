#!/usr/bin/env bash

# 输入参数
# nodeType="$1"
# hardWare="$2"
# host_ip="$3"
tmp_output_file="$4"

# 消息码
# error类
ERROR_KUBELET_NOT_RUN_CODE="error_4001"
ERROR_K8S_NOT_INSTALL_CODE="error_4002"
ERROR_NO_POD_INFO_CODE="error_4003"
ERROR_DOCKER_SERVICE_CODE="error_4004"
ERROR_CGROUP_NOT_MATCH_CODE="error_4005"
ERROR_CALICO_STATUS_CODE="error_4006"

# warning类
WARNING_K8S_COMPONETN_NOT_MATCH_CODE="warning_4001"

# info类
INFO_K8S_VERSION_RECOMMEND_CODE="info_4001"
INFO_PERMISSION_DENIED_CODE="info_4002"

kubelet_version=""
kubectl_version=""
kubeadm_version=""
kubelet_cgroup=""

# 载入公共函数
source ./check_env_utils.sh

# 检测kubelet状态，版本
function check_kubelet_service() {
    kubelet_message_code=""
    kubelet_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubelet_version_command="kubelet --version"
    kubelet_status_command="systemctl status kubelet"

    kubelet_status_str=$(${kubelet_status_command} 2>/dev/null)
    kubelet_status_info=$(echo "${kubelet_status_str}" | grep -E "Active:" \
                                                       | awk -F ':' '{print $2}' \
                                                       | awk -F '\) ' '{print $1}' \
                                                       | sed 's/[ ]*$/)/g' \
                                                       | sed 's/^[ ]*//g')

    if [[ "${kubelet_status_str}" != "" ]]
    then
        if [[ "${kubelet_status_info}" == "${SERVICE_STATUS_RUNNING}" ]]
        then
            # kubelet服务正常
            status=${STATUS_NORMAL}
        else
            # kubelet服务异常
            status=${STATUS_ERROR}
            kubelet_message_code="${ERROR_KUBELET_NOT_RUN_CODE}"
        fi
        # kubelet cgroup
        kubelet_cgroup="$(echo "${kubelet_status_str}" | grep -E '\-\-cgroup-driver=' \
                                                       | awk -F 'cgroup-driver=' '{print $2}' \
                                                       | awk '{print $1}')"
        kubelet_service_status="${status}[${kubelet_status_info}]"

        kubelet_version=$(${kubelet_version_command} 2>/dev/null | awk -F '[" ",]' '{print $2}')
        # 找不到kubelet命令
        if [[ "${kubelet_version}" = "" ]]
        then
            kubelet_message_code="${ERROR_K8S_NOT_INSTALL_CODE}"
        fi
    else
        kubelet_message_code="${ERROR_K8S_NOT_INSTALL_CODE}"
    fi
    write_single_line_to_file "${tmp_output_file}" "kubelet" "${kubelet_service_status}" "${kubelet_version}" "${kubelet_message_code}"
}

# 检测kubeadm状态，版本
function check_kubeadm_service() {
    kubeadm_message_code=""
    kubeadm_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubeadm_version_command="kubeadm version"

    kubeadm_version=$(${kubeadm_version_command} 2>/dev/null | awk -F '[,:]' '{print $7}' \
                                                             | sed -e 's/"//g')
    # 存在kubeadm命令
    if [[ "${kubeadm_version}" != "" ]]
    then
        kubeadm_service_status=${STATUS_NORMAL}
        kubeadm_version="${kubeadm_version}"
    else
        kubeadm_message_code="${ERROR_K8S_NOT_INSTALL_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "kubeadm" "${kubeadm_service_status}" "${kubeadm_version}" "${kubeadm_message_code}"
}

# 检测kubectl状态，版本
function check_kubectl_service() {
    kubectl_message_code=""
    kubectl_service_status=${STATUS_SERVER_NOT_INSTALL}
    kubectl_version_command="kubectl version"

    kubectl_version=$(${UNSET_PROXY} && ${kubectl_version_command} 2>/dev/null | head -n 1 \
                                                                               | awk -F '[,:]' '{print $7}' \
                                                                               | sed -e 's/"//g')
    # 存在kubectl命令
    if [[ "${kubectl_version}" != "" ]]
    then
        kubectl_service_status=${STATUS_NORMAL}
        kubectl_version=${kubectl_version}
    else
        kubectl_message_code="${ERROR_K8S_NOT_INSTALL_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "kubectl" "${kubectl_service_status}" "${kubectl_version}" "${kubectl_message_code}"
}

# 检查三个组件的版本是否匹配
function check_version_match() {
    if [[ "${kubelet_version}" != "${kubeadm_version}" ]] && [[ "${kubelet_version}" != "${kubectl_version}" ]]
    then
        version_match_message_code="${WARNING_K8S_COMPONETN_NOT_MATCH_CODE}"
        match_status="The versions of these components are inconsistent"
    elif [[ "" == "${kubelet_version}" ]] || [[ "" == "${kubeadm_version}" ]] || [[ "" == "${kubectl_version}" ]]
    then
        version_match_message_code="${ERROR_K8S_NOT_INSTALL_CODE}"
        match_status="Some components are not install"
    else
        version_match_message_code="${INFO_K8S_VERSION_RECOMMEND_CODE}"
        match_status="The versions of these components are consistent"
    fi

    write_single_line_to_file "${tmp_output_file}" "kubelet/kubeadm/kubectl version" "${match_status}" "" "${version_match_message_code}"
}

# 检查calico组件
function check_calico_status() {
    calico_message_code=""
    calico_pod_status=""
    all_pods_info="$(${UNSET_PROXY} && ${GET_ALL_PODS_COMMAND} 2>&1)"

    if [[ "" != "$(echo "${all_pods_info}" | grep "was refused")" ]]
    then
        # 没有权限访问k8s
        calico_pod_status=${STATUS_PERMISSION_DENIED}
        calico_message_code="${INFO_PERMISSION_DENIED_CODE}"
    else
        calico_pod_info="$(echo "${all_pods_info}" | grep -E "kube-system.*calico-node.*$(hostname) "))"
        if [[ "" == "${calico_pod_info}" ]]
        then
            calico_message_code="${ERROR_NO_POD_INFO_CODE}"
        else
            ready_status="$(echo "${calico_pod_info}" | awk '{print $3}')"
            run_status="$(echo "${calico_pod_info}" | awk '{print $4}')"
            local status=""
            if [[ "1/1" != "${ready_status}" ]] && [[ "${POD_STATUS_RUNNING}" != "${run_status}" ]]
            then
                status="${STATUS_ERROR}"
                calico_message_code="${ERROR_CALICO_STATUS_CODE}"
            else
                status="${STATUS_NORMAL}"
            fi
            calico_pod_status="${status}[${run_status}]    ${ready_status}"
        fi
    fi

    write_single_line_to_file "${tmp_output_file}" "calico" "${calico_pod_status}" "" "${calico_message_code}"
}

# 检查cgroup是否一致
function check_service_cgroup() {
    cgroup_message_code=""
    cgroup_status=""
    docker_cgroup="$(docker info 2>/dev/null | grep "Cgroup Driver:" | awk -F ':' '{print $2}' | sed 's/ //g')"
    if [[ "" == "${docker_cgroup}" ]]
    then
        cgroup_message_code="${ERROR_DOCKER_SERVICE_CODE}"
        write_single_line_to_file "${tmp_output_file}" "docker/kubelet cgroup-driver" "" "" "${cgroup_message_code}"
    else
        if [[ "${docker_cgroup}" != "${kubelet_cgroup}" ]]
        then
            cgroup_message_code="${ERROR_CGROUP_NOT_MATCH_CODE}"
            cgroup_status="cgroup-driver is inconsistent"
        else
            cgroup_status="cgroup-driver is consistent"
        fi
        write_single_line_to_file "${tmp_output_file}" "docker/kubelet cgroup-driver" "" "" "${cgroup_message_code}"
        write_single_line_to_file "${tmp_output_file}" "" "docker cgroup-driver: ${docker_cgroup}" "" ""
        write_single_line_to_file "${tmp_output_file}" "" "kubelet cgroup-driver: ${kubelet_cgroup}" "" ""
        write_single_line_to_file "${tmp_output_file}" "" "" "" ""
        write_single_line_to_file "${tmp_output_file}" "" "${cgroup_status}" "" ""
    fi

}

function do_check() {
    check_kubelet_service
    check_kubeadm_service
    check_kubectl_service
    check_version_match
    check_calico_status
    check_service_cgroup
}

do_check