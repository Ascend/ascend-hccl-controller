#!/bin/bash
readonly TRUE=1
readonly FALSE=0
readonly kernel_version=$(uname -r)
readonly arch=$(uname -m)
readonly BASE_DIR=$(cd "$(dirname $0)" > /dev/null 2>&1; pwd -P)
readonly PYLIB_PATH=${BASE_DIR}/resources/pylibs

function get_os_name()
{
    local os_name=$(grep -oP "^ID=\"?\K\w+" /etc/os-release)
    echo ${os_name}
}

readonly g_os_name=$(get_os_name)

function get_os_version()
{
    local os_version=$(grep -oP "^VERSION_ID=\"?\K\w+\.?\w*" /etc/os-release)
    echo ${os_version}
    return
}
function install_ansible()
{
    local have_ansible_cmd=$(command -v ansible | wc -l)
    if [[ ${have_ansible_cmd} == 0 ]];then
        export DEBIAN_FRONTEND=noninteractive
        export DEBIAN_PRIORITY=critical
        RESOURCE_DIR=~/resources
        if [ ! -d $RESOURCE_DIR ];then
            echo "error: no resource dir $RESOURCE_DIR"
	        return
        fi
        echo "resource dir=$RESOURCE_DIR"

        echo "dpkg -i --force-all $RESOURCE_DIR/${os_name}_${os_version}_${arch}/python/*.deb"
        dpkg -i --force-all $RESOURCE_DIR/${os_name}_${os_version}_${arch}/python/*.deb
        local python3_version=$(python3 -V)
        if [[ ! "$(python3_version)" =~ "Python 3.6." ]]; then
            echo "python3_version is '$(python3_version)'"
            echo "error: python3 must be python3.6 provided by the system by default, check it by run 'python3 -V'"
	        return
        fi
        python3 -m pip install --upgrade pip --no-index --find-links $RESOURCE_DIR/pylibs
        python3 -m pip install ansible --no-index --find-links $RESOURCE_DIR/pylibs
    else
        echo "ansible is already installed"
    fi
    ansible --version >/dev/null 2>&1
    if [[ $? != 0 ]];then
        echo "error: ansible is not available, check it by run 'ansible --version'"
    fi
}

function main()
{
    local os_name=$(get_os_name)
    local os_version=$(get_os_version)
    echo "OS NAME=${os_name}"
    echo "OS VERSION=${os_version}"
    install_ansible
    if [ ! -d ~/.ansible/roles ];then
        mkdir -p ~/.ansible/roles
    fi
    cp -rf playbooks/roles/* ~/.ansible/roles/
}

main $*
