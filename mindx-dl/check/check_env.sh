# Copyright © Huawei Technologies Co., Ltd. 2020. All rights reserved.
#!/usr/bin/env bash

# 全局变量
STATUS_NORMAL="normal"
STATUS_ERROR="error"
SERVER_NOT_INSTALL="not install"
POD_STATUS_RUNNING="Running"
SERVICE_STATUS_RUNNING="active (running)"
GET_ALL_PODS_COMMAND="kubectl get pods -A -o wide"
DESCRIBE_POD_COMMAND="kubectl describe pod"
NPU_INFO_COMMAND="npu-smi info"
DOCKER_IMAGES="docker images"
MASTER_NODE="master"
WORKER_NODE="worker"
# 既是master又是worker
MASTER_WORKER_NODE="master_worker"

# MindX DL服务名称
DEVICE_PLUGIN_SERVICE_NAME="Ascend-Device-Plugin"
CADVISOR_SERVICE_NAME="cAdvisor"
HCCL_SERVICE_NAME="HCCL-Controller"
VOLCANO_SERVICE_NAME="Volcano"

# 检查的节点类型
if [[ $1 == "${MASTER_NODE}" ]]
then
    CHECK_NODE_TYPE=${MASTER_NODE}
elif [[ $1 == "${WORKER_NODE}" ]]
then
    CHECK_NODE_TYPE=${WORKER_NODE}
else
    CHECK_NODE_TYPE=${MASTER_WORKER_NODE}
fi

# 检查NPU驱动状态和版本
function check_npu() {
    npu_driver_status=${SERVER_NOT_INSTALL}
    npu_driver_version=""
    NPU_INFO_COMMAND="npu-smi info"
    ${NPU_INFO_COMMAND} > /dev/null

    if [ $? == 0 ]
    then
        npu_driver_version=`${NPU_INFO_COMMAND} 2>/dev/null | grep -E 'npu-smi(.*)Version:' \
                                                            | awk -F ':' '{print $2}' \
                                                            | awk -F ' ' '{print $1}' \
                                                            | sed -e 's/[ \t]*$//g' \
                                                            | sed -e 's/^[ \t]*//g'`
        npu_driver_status=${STATUS_NORMAL}
    fi

    echo -e "npu-driver\t"${npu_driver_status}"\t"${npu_driver_version}
}

# 检查docker状态和版本
function check_docker() {
    docker_service_status=${SERVER_NOT_INSTALL}
    docker_version=""
    docker_version_command="docker --version"
    docker_status_command="systemctl status docker"
    # 检查docker是否安装
    docker_status_str=`${docker_status_command} 2>/dev/null | grep -E "Active:" \
                                                            | awk -F ':' '{print $2}' \
                                                            | awk -F 'since' '{print $1}' \
                                                            | sed -e 's/[ \t]*$//g' \
                                                            | sed -e 's/^[ \t]*//g'`
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
    echo -e "docker\t"${docker_service_status}"\t"${docker_version}
}

# 检查K8s状态和版本
function check_kubelet_service() {
    # ========================================检测kubelet状态，版本 =======================
    kubelet_service_status=${SERVER_NOT_INSTALL}
    kubelet_version=""
    kubelet_version_command="kubelet --version"
    kubelet_status_command="systemctl status kubelet"

    kubelet_status_str=`${kubelet_status_command} 2>/dev/null | grep -E "Active:" \
                                                               | awk -F ':' '{print $2}' \
                                                               | awk -F 'since' '{print $1}' \
                                                               | sed 's/[ \t]*$//g' \
                                                               | sed 's/^[ \t]*//g'`

    if [ "${kubelet_status_str}" != "" ]
    then
        if [[ "${kubelet_status_str}" == "${SERVICE_STATUS_RUNNING}" ]]
        then
            # kubeletr服务正常
            status=${STATUS_NORMAL}
        else
            # kubeletr服务异常
            status=${STATUS_ERROR}
        fi
        kubelet_service_status="${status}[$kubelet_status_str]"

        kubelet_version=`${kubelet_version_command} 2>/dev/null | awk -F '[" ",]' '{print $2}'`
        # 找不到kubeletr命令
        if [[ "${kubelet_version}" = "" ]]
        then
            kubelet_version="${STATUS_ERROR}[The kubelet command cannot be found. The kubelet environment may be damaged.]"
        fi
    fi

    echo -e "kubelet\t"${kubelet_service_status}"\t"${kubelet_version}
}

function check_kubeadm_service() {
    # ========================================检测kubeadm状态，版本 =======================
    kubeadm_service_status=${SERVER_NOT_INSTALL}
    kubeadm_version=""
    kubeadm_version_command="kubeadm version"

    kubeadm_version=`${kubeadm_version_command} 2>/dev/null | awk -F '[,:]' '{print $7}' \
                                                            | sed -e 's/"//g'`
    # 找不到kubeadmr命令
    if [[ "${kubeadm_version}" != "" ]]
    then
        kubeadm_service_status=${STATUS_NORMAL}
        kubeadm_version="${kubeadm_version}"
    fi

    echo -e "kubeadm\t"${kubeadm_service_status}"\t"${kubeadm_version}
}

function check_kubectl_service() {
    # ========================================检测kubectl状态，版本 =======================
    kubectl_service_status=${SERVER_NOT_INSTALL}
    kubectl_version=""
    kubectl_version_command="kubectl version"

    kubectl_version=`${kubectl_version_command} 2>/dev/null | head -n 1 \
                                                            | awk -F '[,:]' '{print $7}' \
                                                            | sed -e 's/"//g'`
    # 找不到kubectlr命令
    if [[ "${kubectl_version}" != "" ]]
    then
        kubectl_service_status=${STATUS_NORMAL}
        kubectl_version=${kubectl_version}
    fi

    echo -e "kubectl\t"${kubectl_service_status}"\t"${kubectl_version}
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
    service_status=${SERVER_NOT_INSTALL}
    service_status_str=`${GET_ALL_PODS_COMMAND} 2>/dev/null | grep \`hostname\` \
                                                                  | grep -E "${service_name_regexp}" \
                                                                  | head -n 1 \
                                                                  | sed -e 's/[ \t]*$//g' \
                                                                  | sed -e 's/^[ \t]*//g'`
    # 节点部署了服务
    if [[ "${service_status_str}" != "" ]]
    then

        # 使用的k8s命名空间
        pod_namespace=`echo ${service_status_str} | awk -F ' ' '{print $1}'`
        # pod名称
        pod_name=`echo ${service_status_str} | awk -F ' ' '{print $2}'`
        # pod状态
        pod_status=`echo ${service_status_str} | awk -F ' ' '{print $4}'`
        # 查询pod详情
        describe_pod_command="${DESCRIBE_POD_COMMAND} -n ${pod_namespace} ${pod_name}"
        # pod使用的镜像
        service_image=`${describe_pod_command} 2>/dev/null | grep "Image:" | awk -F ' ' '{print $2}' \
                                                                   | sed -e 's/[ \t]*$//g' \
                                                                   | sed -e 's/^[ \t]*//g'`
        if [[ "${pod_status}" != "${POD_STATUS_RUNNING}" ]]
        then
            status=${STATUS_ERROR}
        else
            status=${STATUS_NORMAL}
        fi
        service_status="${status}[$pod_status]"
    else
        service_image=`${DOCKER_IMAGES} 2>/dev/null | grep -E "${image_name}" | awk -F ' ' '{print $1":"$2}' | xargs | sed -e 's/ /，/g'`
    fi

    echo -e ${service_name}"\t"${service_status}"\t"${service_image}
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
    service_image=`${DOCKER_IMAGES} 2>/dev/null | grep -E "${image_name}" | awk -F ' ' '{print $1":"$2}' | xargs | sed -e 's/ /，/g'`

    if [[ "${service_image}" != "" ]]
    then
        service_status=${STATUS_NORMAL}
    else
        service_status="${STATUS_ERROR}[The image does not exist.]"
    fi

    echo -e ${VOLCANO_SERVICE_NAME}"\t"${service_status}"\t"${service_image}
}

# 检查MindX DL组件状态和版本
function check_mindxdl() {
    # master节点
    if [[ "${CHECK_NODE_TYPE}" =~ ${MASTER_NODE}(.*) ]]
    then
        check_hccl_service
        check_volcano_admission_service
        check_volcano_base_service
        check_volcano_controllers_service
        check_volcano_scheduler_service
    fi
    # worker节点
    if [[ "${CHECK_NODE_TYPE}" =~ (.*)${WORKER_NODE} ]]
    then
        # echo 111 > /dev/null
        check_device_plugin_service
        check_cadvisor_service
    fi

}

#执行检查
function main() {
    check_npu
    check_docker
    check_k8s
    check_mindxdl
}

main