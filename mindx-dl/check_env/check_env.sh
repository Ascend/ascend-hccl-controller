#!/usr/bin/env bash

# 配置检查项，需要检查设置为"yes"，不检查设置为"no"
check_env_os="yes" # 检查操作系统相关内容
check_env_runpkg="yes" # 检查驱动和固件相关内容
check_env_docker="yes" # 检查docker相关内容
check_env_k8s="yes" # 检查Kubernetes相关内容
check_env_mindxdl="yes" # 检查MindX DL组件相关内容

# 输入参数
nodeType=""
hardWare=""
host_ip=""

# 输出文件
current_dir=$(dirname "$(readlink -f "$0")")
tmp_output_file="${current_dir}/tmp_report.txt"
output_file="${current_dir}/env_check_report.txt"

cd "${current_dir}"

source ./check_env_utils.sh

function help_document() {
    echo ""
    echo "    注：配置检查项目需要修改${0}文件"
    echo ""
    echo "    -nt,  --nodetype  (必填) 集群中节点类型，三个可选项 ${MASTER_NODE}, ${WORKER_NODE}, ${MASTER_WORKER_NODE}"
    echo "                             ${MASTER_NODE}：表示管理节点"
    echo "                             ${WORKER_NODE}：表示计算节点"
    echo "                             ${MASTER_WORKER_NODE}：表示既是管理节点也是计算节点"
    echo ""
    echo "    -hw,  --hardware  (必填) 硬件形态，四个可选项 ${HW_COMMON}, ${HW_TRAIN}, ${HW_INFER}, ${HW_300T}"
    echo "                             ${HW_COMMON}：表示通用服务器，且只能在 -nt 参数的值为${MASTER_NODE}时使用"
    echo "                             ${HW_TRAIN}：表示Atlas 800训练服务器"
    echo "                             ${HW_INFER}：表示Atlas 800推理服务器"
    echo "                             ${HW_300T}：表示服务器插Atlas 300T训练卡"
    echo ""
    echo "    -ip               (选填) 本机ip，如：10.10.123.123"
    echo "                             如果没有该参数，输出结果不会包含本机ip的信息"
    echo ""
    echo "    -h,   --help             显示帮助信息"
    echo ""
    echo "    示例1，本节点是Atlas 800训练服务器，作为集群中的计算节点"
    echo "        bash check_env.sh -nt ${WORKER_NODE} -hw ${HW_TRAIN}"
    echo ""
    echo "    示例2，本节点是通用服务器，在集群中是管理节点"
    echo "        bash check_env.sh -nt ${MASTER_NODE} -hw ${HW_COMMON} -ip 10.10.123.123"
    echo ""
    exit 0
}

# 检查节点类型输入参数
function check_param_nodetype() {
    if [[ "" == "${nodeType}" ]]
    then
        echo -e "\n-nt参数不能为空!\n"
        exit 1
    fi

    if [[ "${MASTER_NODE}" != "${nodeType}" ]] && \
        [[ "${WORKER_NODE}" != "${nodeType}" ]] && \
        [[ "${MASTER_WORKER_NODE}" != "${nodeType}" ]]
    then
        echo -e "\n-nt参数只能为${MASTER_NODE}, ${WORKER_NODE}, ${MASTER_WORKER_NODE}其中之一\n"
        exit 1
    fi
}

# 检查硬件形态输入参数
function check_param_hardware() {
    if [[ "" == "${hardWare}" ]]
    then
        echo -e "\n-hw参数不能为空!\n"
        exit 1
    fi

    if [[ "${HW_TRAIN}" != "${hardWare}" ]] && \
        [[ "${HW_INFER}" != "${hardWare}" ]] && \
        [[ "${HW_300T}" != "${hardWare}" ]] && \
        [[ "${HW_COMMON}" != "${hardWare}" ]]
    then
        echo -e "\n-hw参数只能为${HW_COMMON}, ${HW_TRAIN}, ${HW_INFER}, ${HW_300T}其中之一\n"
        exit 1
    fi

    if [[ "${MASTER_NODE}" == "${nodeType}" && "${HW_COMMON}" != "${hardWare}" ]] || \
        [[ "${MASTER_NODE}" != "${nodeType}" && "${HW_COMMON}" == "${hardWare}" ]]
    then
        echo -e "\n-hw(--hardware) ${HW_COMMON}和 -nt(--nodetype) ${MASTER_NODE}为固定搭配\n"
        exit 1
    fi
}

# 检查ip输入参数
function check_param_ip() {
    local ip_regexp="^([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])$"
    if [[ ! "${host_ip}" =~ ${ip_regexp} ]] &&\
        [[ "${host_ip}" !=  "" ]]
    then
        echo -e "\n${host_ip}不是一个有效ip地址\n"
        exit 1
    fi
}

# 输出主机相关信息
function print_host_info() {
    sed -i "1i hostname: $(hostname)" "${tmp_output_file}"
    if [[ "" != "${host_ip}" ]]
    then
        sed -i "1i ip: ${host_ip}" "${tmp_output_file}"
    fi
}

function check_input_params() {
    check_param_nodetype
    check_param_hardware
    check_param_ip
}

function to_report_file() {
    print_host_info
    # 表头
    sed -iE '/^hostname/a service*|*status*|*version*|*message code*|' "${tmp_output_file}"
    # 表格格式化
    cat "${tmp_output_file}" | column -s "*" -t >> "${output_file}"
    # 添加行分隔符
    row_max_length=$(wc -L "${output_file}" | awk '{print $1}')
    split_line_symbel=$(printf "%-${row_max_length}s\n" "-" | sed -e 's/ /-/g')
    sed -iE "/^[A-Za-z]/i ${split_line_symbel}" "${output_file}"
    # 表尾
    echo "${split_line_symbel}" >> "${output_file}"

    chmod 400 "${output_file}"
}

function to_officia_report() {
    # 有新的检查，先删除旧的检查报告
    if [[ "${check_env_os}${check_env_runpkg}${check_env_k8s}${check_env_docker}${check_env_mindxdl}" =~ yes ]] &&
        [[ -e "${tmp_output_file}" ]]
    then
        rm -f "${output_file}"
        rm -f "${output_file}E"
        # 将临时报告转换成正式报告
        to_report_file >/dev/null 2>&1
        rm -f "${tmp_output_file}"
        rm -f "${tmp_output_file}E"
        rm -f "${output_file}E"
        echo -e "\nFinished! The check report is stored in ${output_file}\n"
    fi
}

function execute_check() {
    # 参数检查
    check_input_params

     rm -f "${tmp_output_file}"
     rm -f "${tmp_output_file}E"

    # 检查操作系统相关内容
    if [[ -e "${current_dir}/check_env_os.sh" ]] && [[ "yes" == "${check_env_os}" ]]
    then
        bash check_env_os.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查驱动和固件相关内容
    if [[ -e "${current_dir}/check_env_runpkg.sh" ]] && [[ "yes" == "${check_env_runpkg}" ]]
    then
        bash check_env_runpkg.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查docker相关内容
    if [[ -e "${current_dir}/check_env_docker.sh" ]] && [[ "yes" == "${check_env_docker}" ]]
    then
        bash check_env_docker.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查Kubernetes相关内容
    if [[ -e "${current_dir}/check_env_k8s.sh" ]] && [[ "yes" == "${check_env_k8s}" ]]
    then
        bash check_env_k8s.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查MindX DL组件相关内容
    if [[ -e "${current_dir}/check_env_mindxdl.sh" ]] && [[ "yes" == "${check_env_mindxdl}" ]]
    then
        bash check_env_mindxdl.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 生成正式报告
    to_officia_report
}

while [ -n "$1" ]
do
  case "$1" in
    -nt|--nodetype)
        nodeType=$2;
        shift
        ;;
    -hw|--hardware)
        hardWare=$2
        shift
        ;;
    -ip)
        host_ip=$2
        shift
        ;;
    -h|--help)
        help_document;
        exit
        ;;
    *)
        echo "$1不是一个有效的选项，请使用--help"
        exit 1
        ;;
  esac
  shift
done

execute_check