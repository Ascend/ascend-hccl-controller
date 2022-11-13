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

function get_os_version()
{
    local os_version=$(grep -oP "^VERSION_ID=\"?\K\w+\.?\w*" /etc/os-release)
    echo ${os_version}
}

function install_ansible()
{
    local have_ansible_cmd=$(command -v ansible | wc -l)
    if [[ ${have_ansible_cmd} == 0 ]];then
        RESOURCE_DIR=~/resources
        if [ ! -d $RESOURCE_DIR ];then
            echo "error: no resource dir $RESOURCE_DIR"
	        return 1
        fi
        echo "resource dir=$RESOURCE_DIR"

        local have_rpm=0
        case ${os_name} in
        centos|openEuler|kylin)
            local have_rpm=1
            ;;
        ubuntu)
            local have_rpm=0
            ;;
        *)
            echo "error: check OS ${os_name} fail"
            return 1
            ;;
        esac

        if [[ ${have_rpm} == 1 ]] && [[ "${os_name}" != "openEuler" ]]; then
            rpm -ivh --force --nodeps --replacepkgs $RESOURCE_DIR/${os_name}_${os_version}_${arch}/python/*.rpm
        fi
        if [[ ${have_rpm} == 0 ]]; then
            export DEBIAN_FRONTEND=noninteractive && export DEBIAN_PRIORITY=critical; dpkg --force-all -i $RESOURCE_DIR/${os_name}_${os_version}_${arch}/python/*.deb
        fi

        local python3_version=$(python3 -V)
        echo "python3_version is '${python3_version}'"
        case ${os_name} in
        centos)
            if [[ ! "${python3_version}" =~ "Python 3.6." ]]; then
                echo "error: python3 must be Python 3.6 provided by the centos 7.6 by default, check it by run 'python3 -V'"
                return 1
            fi
            ;;
        ubuntu)
            case ${os_version} in
            18.04)
                if [[ ! "${python3_version}" =~ "Python 3.6." ]]; then
                echo "error: python3 must be Python 3.6 provided by the ubuntu 18.04 by default, check it by run 'python3 -V'"
                return 1
                fi
                ;;
            20.04)
                if [[ ! "${python3_version}" =~ "Python 3.8." ]]; then
                echo "error: python3 must be Python 3.8 provided by the ubuntu 20.04 by default, check it by run 'python3 -V'"
                return 1
                fi
                ;;
            *)
                echo "error: check version ${os_version} fail"
                return 1
                ;;
            esac
            ;;
        openEuler|kylin)
            if [[ ! "${python3_version}" =~ "Python 3.7." ]]; then
                echo "error: python3 must be Python 3.7 provided by the openEuler 20.03 or kylin v10 sp1 by default, check it by run 'python3 -V'"
                return 1
            fi
            ;;
        *)
            echo "error: check OS ${os_name} fail"
            return 1
            ;;
        esac

        python3 -m pip install --upgrade pip --no-index --find-links $RESOURCE_DIR/pylibs
        python3 -m pip install ansible --no-index --find-links $RESOURCE_DIR/pylibs
    else
        echo "ansible is already installed"
    fi
    ansible --version >/dev/null 2>&1
    if [[ $? != 0 ]];then
        echo "error: ansible is not available, check it by run 'ansible --version'"
        return 1
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
