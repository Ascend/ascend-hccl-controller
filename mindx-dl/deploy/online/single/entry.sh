#!/bin/bash
# Entry script for online deployment of a single node
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e
ansible-playbook -vv docker_install.yaml
ansible-playbook -vv install.yaml
ansible-playbook -vv deploy.yaml