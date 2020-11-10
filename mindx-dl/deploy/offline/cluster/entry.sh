#!/bin/bash
# Entry script for offline cluster deployment.
# Copyright © Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e
ansible-playbook -vv set_global_env.yaml
ansible-playbook -vv offline_install_package.yaml
ansible-playbook -vv offline_load_images.yaml
ansible-playbook -vv init_kubernetes.yaml
ansible-playbook -vv clean_services.yaml
ansible-playbook -vv offline_deploy_service.yaml