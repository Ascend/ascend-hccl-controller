#!/bin/bash
# Entry script for online cluster deployment.
# Copyright © Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e

# kubernetes, nfs, go, and docker are installed by default.
# SKIP_INSTALL_SOFTWARE="kubernetes" refer to install nfs, go, and docker,but don't install kubernetes
SKIP_INSTALL_SOFTWARE=""

# two options: no_asend-hub/asend-hub
GET_MINDXDL_IMAGE_METHOD="ascend-hub"

ansible-playbook -vv set_global_env.yaml
ansible-playbook -vv --skip-tags=${SKIP_INSTALL_SOFTWARE} online_install_package.yaml
ansible-playbook -vv --tags=${GET_MINDXDL_IMAGE_METHOD} online_load_images.yaml
ansible-playbook -vv init_kubernetes.yaml
ansible-playbook -vv clean_services.yaml
ansible-playbook -vv online_deploy_service.yaml