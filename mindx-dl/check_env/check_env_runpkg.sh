#!/usr/bin/env bash

# 输入参数
nodeType="$1"
hardWare="$2"
# host_ip="$3"
tmp_output_file="$4"

# 消息码
# error类
ERROR_DRIVER_NOT_INSTALL_CODE="error_2001"
ERROR_DEVICE_IP_DUPLICATE_CODE="error_2002"
ERROR_CANNOT_FIND_UPGRADE_TOOL_CODE="error_2003"
ERROR_NO_HCCN_TOOL_CODE="error_2004"
ERROR_FIRMWARE_NOT_INSTALL_CODE="error_2005"
ERROR_NO_INSTALL_PATH_FILE_CODE="error_2006"

# warning类
WARNING_DEVICE_LINK_DOWN_CODE="warning_2001"

# info类
INFO_NO_NUP_NUM_CODE="info_2001"
INFO_NO_DEVICE_IP_FILE_CODE="info_2002"
INFO_DEVICE_LINK_EMPTY_CODE="info_2003"
INFO_DEVICE_PKG_EMPTY_CODE="info_2004"

# firmware安装路径
firmware_install_path="$(cat /etc/ascend_install.info 2>/dev/null | grep 'Firmware_Install_Path_Param' \
                                                                  | awk -F '=' '{print $2}')"
# driver安装路径
driver_install_path="$(cat /etc/ascend_install.info 2>/dev/null | grep 'Driver_Install_Path_Param' \
                                                                | awk -F '=' '{print $2}')"
# driver安装状态
npu_driver_status="${STATUS_SERVER_NOT_INSTALL}"
# npu数量
npu_num=""
# 设备相关信息
device_info=""
hccn_tool_error="false"
upgrade_tool_error="false"

# 载入公共函数
source ./check_env_utils.sh

# 检查工具是否可用
function check_hccn_tool() {
    hccn_content="$(hccn_tool --help 2>/dev/null)"
    if [[ "" == "${hccn_content}" ]]
    then
        hccn_tool_error='true'
    fi
}

# 检查工具是否可用
function check_upgrade_tool() {
    device_info="$(${driver_install_path}/driver/tools/upgrade-tool --device_index -1 --system_version 2>/dev/null)"
    if [[ "" == "${device_info}" ]]
    then
        upgrade_tool_error='true'
    fi
}

# 检查driver版本
function check_driver_version() {
    driver_message_code=""
    npu_driver_version="$(npu-smi info 2>/dev/null | grep -E 'npu-smi(.*)Version:' \
                                                   | awk -F ':' '{print $2}' \
                                                   | awk -F ' ' '{print $1}' \
                                                   | sed -e 's/[ ]*$//g' \
                                                   | sed -e 's/^[ ]*//g')"
    if [[ "" != "${npu_driver_version}" ]]
    then
        npu_driver_status=${STATUS_INSTALLED}
    else
        npu_driver_status=${STATUS_SERVER_NOT_INSTALL}
        driver_message_code="${ERROR_DRIVER_NOT_INSTALL_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "npu-driver" "${npu_driver_status}" "${npu_driver_version}" "${driver_message_code}"
}

# 检查driver数量
function check_npu_number() {
    npu_message_code=""

    if [[ "${npu_driver_status}" == "${STATUS_INSTALLED}" ]]
    then
        if [[ "true" != "${upgrade_tool_error}" ]]
        then
            # upgrade-tool工具
            npu_num="$(echo "${device_info}" | grep 'deviceId' | wc -l)"
            if [[ "" == "${npu_num}" ]]
            then
                npu_message_code="${INFO_NO_NUP_NUM_CODE}"
            fi
        else
            npu_message_code="${ERROR_CANNOT_FIND_UPGRADE_TOOL_CODE}"
        fi
    else
        npu_message_code="${ERROR_DRIVER_NOT_INSTALL_CODE}"
    fi

    write_single_line_to_file "${tmp_output_file}" "number of npu" "${npu_num}" "" "${npu_message_code}"
}

# 检查device_ip
function check_device_ip() {
    ips_message_code=" "
    duplicate_message_code=""
    duplicate_line="Device ip not configured"
    device_ips="$(cat /etc/hccn.conf | grep 'address' \
                                     | sort \
                                     | sed 's/address/device/g' \
                                     | sed 's/^/ *|* *|*/g' \
                                     | sed 's/$/*|* *|* *|/g')"
    if [[ "" == "${device_ips}" ]]
    then
        ips_message_code="${INFO_NO_DEVICE_IP_FILE_CODE}"
        duplicate_message_code="${INFO_NO_DEVICE_IP_FILE_CODE}"
    else
        # ip是否配置重复
        duplicate_line="$(echo "${device_ips}" | awk -F "=" '{print $2} | uniq -d')"
        if [[ "" != "${duplicate_line}" ]]
        then
            duplicate_message_code="${ERROR_DEVICE_IP_DUPLICATE_CODE}"
        else
            duplicate_line="No duplicate device ip"
        fi
    fi

    local content_ips=" *|*device_ip configure info*|* *|* *|*${ips_message_code}*|\n${device_ips}"
    write_multiline_to_file "${tmp_output_file}" "${content_ips}"
    write_single_line_to_file "${tmp_output_file}" "duplicate device_ip" "${duplicate_line}" "" "${duplicate_message_code}"
}

# 检查设备网口状态
function check_device_link() {

    link_message_code=" "
    link_content=""
    if [[ "true" == "${hccn_tool_error}" ]]
    then
        link_message_code="${ERROR_NO_HCCN_TOOL_CODE}"
    else
        device_list="$(cat /etc/hccn.conf | grep 'address' | sort | awk -F '[_=]' '{print $2}' | xargs)"
        if [[ "" == "${device_list}" ]]
        then
            link_message_code="${INFO_NO_DEVICE_IP_FILE_CODE}"
        else
            for i in ${device_list}
            do
                status_message_code=" "
                link_status="$(hccn_tool -i ${i} -link -g 2>/dev/null | grep 'link status' \
                                                                      | awk -F ":" '{print $2}' \
                                                                      | sed 's/ //g')"
                if [[ "" == "${link_status}}" ]]
                then
                    status_message_code="${INFO_DEVICE_LINK_EMPTY_CODE}"
                elif [[ "${LINK_STATUS_DOWN}" == "${link_status}" ]]
                then
                    status_message_code="${WARNING_DEVICE_LINK_DOWN_CODE}"
                fi
                link_content="${link_content}\n *|* *|*device_${i}: ${link_status}*|* *|*${status_message_code}*|"
            done
        fi
    fi

    local content=" *|*device link status*|* *|* *|*${link_message_code}*|${link_content}"
    write_multiline_to_file "${tmp_output_file}" "${content}"
}

# 检查设备收发包情况
function check_device_packages() {
    pkg_message_code=" "
    pkg_content=""
    if [[ "true" == "${hccn_tool_error}" ]]
    then
        pkg_message_code="${ERROR_NO_HCCN_TOOL_CODE}"
    else
        device_list="$(cat /etc/hccn.conf | grep 'address' | sort | awk -F '[_=]' '{print $2}' | xargs)"
        if [[ "" == "${device_list}" ]]
        then
            pkg_message_code="${INFO_NO_DEVICE_IP_FILE_CODE}"
        else
            for i in ${device_list}
            do
                pkg_data_message_code=" "
                pkg_data="$(hccn_tool -i ${i} -stat -g 2>/dev/null | grep -E 'mac_tx_total_pkt_num|mac_rx_total_pkt_num|mac_tx_bad_pkt_num|mac_rx_bad_pkt_num')"
                if [[ "" == "${pkg_data}}" ]]
                then
                    pkg_data_message_code="${INFO_DEVICE_PKG_EMPTY_CODE}"
                    pkg_content="${pkg_content}\n *|* *|*device_${i}: *|* *|*${pkg_data_message_code}*|"
                else
                    pkg_data="$(echo "${pkg_data}" | sed 's/^/ *|* *|*/g' | sed 's/$/*|* *|* *|/g')"
                    pkg_content="${pkg_content}\n *|* *|*device_${i}:*|* *|* *|\n${pkg_data}"
                    pkg_content="${pkg_content}\n *|* *|* *|* *|* *|"
                fi
            done
        fi
    fi

    local content=" *|*statistics of device's packages*|* *|* *|*${pkg_message_code}*|${pkg_content}"
    write_multiline_to_file "${tmp_output_file}" "${content}"
}

function check_firmware() {
    firmware_message_code=" "
    firmware_status="${STATUS_SERVER_NOT_INSTALL}"
    firmware_version=""
    if [[ "" == "${firmware_install_path}" ]]
    then
        firmware_message_code="${ERROR_NO_INSTALL_PATH_FILE_CODE}"
    else
        firmware_version="$(cat "${firmware_install_path}/firmware/version.info" 2>/dev/null | grep "Version=" | awk -F "=" '{print $2}')"

        if [[ "" != "${firmware_version}" ]]
        then
            firmware_status="${STATUS_INSTALLED}"
        else
            firmware_message_code="${ERROR_FIRMWARE_NOT_INSTALL_CODE}"
        fi
    fi

    write_single_line_to_file "${tmp_output_file}" "firmware" "${firmware_status}" "${firmware_version}" "${firmware_message_code}"
}

function do_check() {
    write_category "drive/firmware-related"
    check_upgrade_tool
    check_firmware
    check_driver_version
    check_npu_number
    # 推理不涉及
    if [[ "${HW_TRAIN}" == "${hardWare}" ]] || \
        [[ "${HW_300T}" == "${hardWare}" ]]
    then
        check_hccn_tool
        check_device_ip
        check_device_link
        check_device_packages
    fi
}

# master节点不执行run包相关检查
if [[ "${MASTER_NODE}" != "${nodeType}" ]]
then
    do_check
fi