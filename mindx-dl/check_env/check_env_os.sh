#!/usr/bin/env bash

# 输入参数
# nodeType="$1"
# hardWare="$2"
# host_ip="$3"
tmp_output_file="$4"

# 消息码
# error类
ERROR_USER_NOT_EXISTS_CODE='error_1001'
ERROR_USER_ID_ERROR_CODE='error_1002'
ERROR_NOT_JOIN_GROUP_CODE='error_1003'
ERROR_SWAP_NOT_EMPTY_CODE="error_1004"

# info类
INFO_NOT_SUPPORT_OS_CODE="info_1001"
INFO_CLUSTER_DATE_CODE="info_1002"

# 常量
HWHIAIUSER_ID="1000"
HWMINDX_ID="9000"
EMPTY_SWAP_SIZE="0B"

# 载入公共函数
source ./check_env_utils.sh

function check_os_arch() {
    write_single_line_to_file "${tmp_output_file}" "architecture" "$(arch)"
}

# 逻辑cpu个数
function check_logical_cpu() {
    logic_cpu_count="$(cat /proc/cpuinfo| grep "processor" | wc -l 2>/dev/null)"

    write_single_line_to_file "${tmp_output_file}" "number of logical cpu" "${logic_cpu_count}"
}

# cpu使用率
function check_cpu_utilization() {
    idle_rate="$(top -n 1 2>/dev/null | grep -i '%cpu' | head -n 1 | awk '{print $8}')"
    usage=""
    if [[ "" != "${idle_rate}" ]]
    then
        usage=$(echo "100  ${idle_rate}" | awk '{print $1 - $2}')
    fi

    write_single_line_to_file "${tmp_output_file}" "cpu usage" "${usage}%"
}

# 可用内存
function check_available_memery() {
    available_mem="$(free -h | grep 'Mem:' | awk '{print $7}')"
    write_single_line_to_file "${tmp_output_file}" "available memery" "${available_mem}"
}

# 硬盘大小，使用率
function check_disk_usage() {
    disk_usage_result="$(df -Th | grep -E '^/dev|Filesystem' | awk '{print $1"|"$2"|"$3"|"$4"|"$5"|"$6"|"$7}' \
                                                             | column -s '|' -t \
                                                             | sed 's/^/ *|* *|*/g' \
                                                             | sed 's/$/*|* *|* *|/g')"
    local content=" *|*disk usage*|* *|* *|* *|\n${disk_usage_result}"
    write_multiline_to_file "${tmp_output_file}" "${content}"
}

# swap空间
function check_swap_usage() {
    swap_message_code=" "
    swap_usage_content="$(free -h | grep -E 'Swap: |total' | sed 's/^/ *|* *|*/g' | sed 's/$/*|* *|* *|/g')"
    swap_size="$(echo "${swap_usage_content}" | grep "Swap:" | awk '{print $3}')"
    if [[ "${EMPTY_SWAP_SIZE}" != "${swap_size}" ]]
    then
        swap_message_code="${ERROR_SWAP_NOT_EMPTY_CODE}"
    fi
    local content=" *|*swap usage*|* *|* *|*${swap_message_code}*|\n${swap_usage_content}"
    write_multiline_to_file "${tmp_output_file}" "${content}"
}

# 防火墙
function check_firwall_status() {
    firewall_status=""
    message_code=""
    os_name="$(cat /etc/*release 2>/dev/null | grep -E '^ID=' | awk -F '=' '{print $2}' | sed 's/\"//g')"
    if [[ "centos" == "${os_name}" ]]
    then
        firewall_status="$(systemctl status firewalld.service 2>/dev/null | grep 'Active:' | awk -F ': ' '{print $2}')"
    elif [[ "ubuntu" == "${os_name}" ]]
    then
        firewall_status="$(ufw status 2>/dev/null | grep 'Status:' | awk -F ': ' '{print $2}')"
    fi

    if [[ "" == "${firewall_status}" ]]
    then
        message_code="${INFO_NOT_SUPPORT_OS_CODE}"
    fi
    write_single_line_to_file "${tmp_output_file}" "firewall status" "${firewall_status}" "" "${message_code}"
}

# 检查用户
function check_os_user() {
    hwhiaiuser_message_code=""
    hwmindx_message_code=""

    # 检查HwHiAiUser
    hwhiaiuser_uid="$(id HwHiAiUser 2>/dev/null | awk '{print $1}' | awk -F '=' '{print $2}' | awk -F '(' '{print $1}')"
    hwhiaiuser_gid="$(id HwHiAiUser 2>/dev/null | awk '{print $2}' | awk -F '=' '{print $2}' | awk -F '(' '{print $1}')"

    # HwHiAiUser不存在
    if [[ "" == "${hwhiaiuser_uid}" ]] || [[ "" == "${hwhiaiuser_gid}" ]]
    then
        hwhiaiuser_content="User 'HwHiAiUser' not exists"
        hwhiaiuser_message_code="${ERROR_USER_NOT_EXISTS_CODE}"
    else
        if [[ "${HWHIAIUSER_ID}" != "${hwhiaiuser_uid}" ]] || [[ "${HWHIAIUSER_ID}" != "${hwhiaiuser_gid}" ]]
        then
        # HwHiAiUser gid和uid不为1000
            hwhiaiuser_message_code="${ERROR_USER_ID_ERROR_CODE}"
        fi
        hwhiaiuser_content="HwHiAiUser(uid:${hwhiaiuser_uid}, gid:${hwhiaiuser_gid})"
    fi

    # 检查hwMindX
    hwmindx_uid="$(id hwMindX 2>/dev/null | awk '{print $1}' | awk -F '=' '{print $2}' | awk -F '(' '{print $1}')"
    hwmindx_gid="$(id hwMindX 2>/dev/null | awk '{print $2}' | awk -F '=' '{print $2}' | awk -F '(' '{print $1}')"

    # hwMindX不存在
    if [[ "" == "${hwmindx_uid}" ]] || [[ "" == "${hwmindx_gid}" ]]
    then
        hwmindx_content="User 'hwMindX' not exists"
        hwmindx_message_code="${ERROR_USER_NOT_EXISTS_CODE}"
    else
        if [[ "${HWMINDX_ID}" != "${hwmindx_uid}" ]] || [[ "${HWMINDX_ID}" != "${hwmindx_gid}" ]]
        then
            # hwMindX gid和uid不为9000
            hwmindx_message_code="${ERROR_USER_ID_ERROR_CODE}"
        else
            # hwMindX是否加入HwHiAiUser的用户组
            join_group_str="$(id hwMindX | grep -E 'groups=.*1000\(HwHiAiUser\)')"
            if [[ "" == "${join_group_str}" ]]
            then
                hwmindx_message_code="${ERROR_NOT_JOIN_GROUP_CODE}"
            fi
        fi
        hwmindx_content="hwMindX(uid:${hwmindx_uid}, gid:${hwmindx_gid})"
    fi

    write_single_line_to_file "${tmp_output_file}" "user"
    write_single_line_to_file "${tmp_output_file}" "    " "${hwhiaiuser_content}" "" "${hwhiaiuser_message_code}"
    write_single_line_to_file "${tmp_output_file}" "    " "${hwmindx_content}" "" "${hwmindx_message_code}"
}

function check_date() {
    date_str="$(date +"%A %F %T %Z")"
    write_single_line_to_file "${tmp_output_file}" "date" "${date_str}" "" "${INFO_CLUSTER_DATE_CODE}"
}

function do_check() {
    write_category "os-related"
    check_os_arch
    check_firwall_status
    check_logical_cpu
    check_cpu_utilization
    check_available_memery
    check_date
    check_os_user
    check_disk_usage
    check_swap_usage
}

do_check