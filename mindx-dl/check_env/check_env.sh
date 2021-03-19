#!/usr/bin/env bash

# 配置检查项，需要检查设置为"yes"，不检查设置为"no"
check_env_os="yes" # 检查操作系统相关内容
check_env_driver_firmware="yes" # 检查驱动和固件相关内容
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
    echo "Note: The ${0} file needs to be modified for configure check items."
    echo ""
    echo "  -nt, --nodetype  (required) Node type in cluster."
    echo "                              All options: '${MASTER_NODE}', '${WORKER_NODE}', '${MASTER_WORKER_NODE}'. "
    echo "                                ${MASTER_NODE}: management node"
    echo "                                ${WORKER_NODE}: compute node"
    echo "                                ${MASTER_WORKER_NODE}: this node is both a management node and a compute node."
    echo ""
    echo "  -hw, --hardware  (required) Hardware form."
    echo "                              All options: '${HW_COMMON}', '${HW_TRAIN}', '${HW_INFER}', '${HW_300T}', '${HW_300I_PRO}'"
    echo "                                ${HW_COMMON}: Common server. This parameter can be used"
    echo "                                        only when the '-nt' parameter is set to '${MASTER_NODE}'"
    echo "                                ${HW_TRAIN}: Atlas 800 training servers"
    echo "                                ${HW_INFER}: Atlas 800 inference servers or servers with Atlas 300I(model 3000/3010) inference cards"
    echo "                                ${HW_300T}: servers(with Atlas 300T training cards)"
    echo "                                ${HW_300I_PRO}: servers(with Atlas 300I Pro inference cards)"
    echo ""
    echo "  -ip               (optional) local IP. For example: 10.10.123.123"
    echo "                               If there is no such parameter, the output will not contain"
    echo "                                 the information of local IP"
    echo ""
    echo "  -h,  --help    This command's help information"
    echo ""
    echo "Example 1: the node is an Atlas 800 training server, which acts as the compute node in the cluster."
    echo "           bash check_env.sh -nt ${WORKER_NODE} -hw ${HW_TRAIN}"
    echo ""
    echo "Example 2: this node is a common server and a management node in the cluster."
    echo "           bash check_env.sh -nt ${MASTER_NODE} -hw ${HW_COMMON} -ip 10.10.123.123"
    echo ""
    exit 0
}

# 检查节点类型输入参数
function check_param_nodetype() {
    if [[ "" == "${nodeType}" ]]
    then
        echo -e "\n'-nt(--nodetype)' cannot be empty!\n"
        exit 1
    fi

    if [[ "${MASTER_NODE}" != "${nodeType}" ]] && \
        [[ "${WORKER_NODE}" != "${nodeType}" ]] && \
        [[ "${MASTER_WORKER_NODE}" != "${nodeType}" ]]
    then
        echo -e "\n'-nt(--nodetype)' can only be set to '${MASTER_NODE}', '${WORKER_NODE}', '${MASTER_WORKER_NODE}'.\n"
        exit 1
    fi
}

# 检查硬件形态输入参数
function check_param_hardware() {
    if [[ "" == "${hardWare}" ]]
    then
        echo -e "\n'-hw(--hardware)' cannot be empty!\n"
        exit 1
    fi

    # 硬件形态不在规定范围内
    is_valid='false'
    for i in ${HW_ARR[@]}
    do
        if [[ "$i" == "${hardWare}" ]]
        then
            is_valid="true"
            break
        fi
    done

    if [[ ${is_valid} == 'false' ]]
    then
        echo -e "\n'-hw(--hardware)' can only be set to '${HW_COMMON}', '${HW_TRAIN}', '${HW_INFER}', '${HW_300T}', '${HW_300I_PRO}'.\n"
        exit 1
    fi

    if [[ "${MASTER_NODE}" == "${nodeType}" && "${HW_COMMON}" != "${hardWare}" ]] || \
        [[ "${MASTER_NODE}" != "${nodeType}" && "${HW_COMMON}" == "${hardWare}" ]]
    then
        echo -e "\n'-hw(--hardware) ${HW_COMMON}' and '-nt(--nodetype) ${MASTER_NODE}' are fixed collocations\n"
        exit 1
    fi
}

# 检查ip输入参数
function check_param_ip() {
    local ip_regexp="^([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])$"
    if [[ ! "${host_ip}" =~ ${ip_regexp} ]] &&\
        [[ "${host_ip}" !=  "" ]]
    then
        echo -e "\n'${host_ip}' is not a valid IP address\n"
        exit 1
    fi
}

# 输出主机相关信息
function print_host_info() {
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
    # 表头
    sed -iE '1i Category*|*Check Items*|*Status*|*Version*|*Message Code*|' "${tmp_output_file}"

    print_host_info

    # 表格格式化
    cat "${tmp_output_file}" | column -s "*" -t >> "${output_file}"
    # 添加类别分隔符
    row_max_length=$(wc -L "${output_file}" | awk '{print $1}')
    split_category_symbol=$(printf "%-${row_max_length}s\n" "-" | sed -e 's/ /-/g')
    sed -iE "/^[A-Za-z]/i ${split_category_symbol}" "${output_file}"

    # 空格开头的行添加分隔符
    # 去掉开头的空格，该行长度
    no_space_line_max_length=$(grep -E '^[ ]' "${output_file}" | head -n 1 | sed 's/^\s*//g' | wc -L)
    # 以空格开头的行，空格长度
    space_length=$(( ${row_max_length} - ${no_space_line_max_length} ))
    no_space_symbol=$(printf "%-${no_space_line_max_length}s\n" "-" | sed -e 's/ /-/g')
    space_symbol=$(printf "%${space_length}s")
    item_split_line_symbol="$(echo "${space_symbol}${no_space_symbol}")"
    sed -Ei "/^${space_symbol}\|\s{2}[A-Za-z]\+*/i\\${item_split_line_symbol}" "${output_file}"
    # 表尾
    echo "${split_category_symbol}" >> "${output_file}"

    chmod 400 "${output_file}"
}

function to_officia_report() {
    # 有新的检查，先删除旧的检查报告
    if [[ "${check_env_os}${check_env_driver_firmware}${check_env_k8s}${check_env_docker}${check_env_mindxdl}" =~ yes ]] &&
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
    if [[ "yes" == "${check_env_os}" ]] && [[ -e "${current_dir}/check_env_os.sh" ]]
    then
        bash check_env_os.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查驱动和固件相关内容
    if [[ "yes" == "${check_env_driver_firmware}" ]] && [[ -e "${current_dir}/check_env_runpkg.sh" ]]
    then
        bash check_env_runpkg.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查docker相关内容
    if [[ "yes" == "${check_env_docker}" ]] && [[ -e "${current_dir}/check_env_docker.sh" ]]
    then
        bash check_env_docker.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查Kubernetes相关内容
    if [[ "yes" == "${check_env_k8s}" ]] && [[ -e "${current_dir}/check_env_k8s.sh" ]]
    then
        bash check_env_k8s.sh "$nodeType" "$hardWare" "$host_ip" "${tmp_output_file}" >/dev/null 2>&1
    fi

    # 检查MindX DL组件相关内容
    if [[ "yes" == "${check_env_mindxdl}" ]] && [[ -e "${current_dir}/check_env_mindxdl.sh" ]]
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
        echo "'$1' is not a valid option, please use --help or -h."
        exit 1
        ;;
  esac
  shift
done

execute_check