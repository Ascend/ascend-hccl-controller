#!/bin/bash
# Entry script for offline cluster deployment.
# Copyright © Huawei Technologies Co., Ltd. 2020. All rights reserved.
set -e

scope="basic"

if [ $scope == "basic" ]
then
  ansible-playbook -vv set_global_env.yaml --tags=basic_only
  ansible-playbook -vv offline_install_packages.yaml --tags=basic_only
  ansible-playbook -vv offline_load_images.yaml --tags=basic_only
  ansible-playbook -vv init_kubernetes.yaml --tags=basic_only
  ansible-playbook -vv clean_services.yaml
  ansible-playbook -vv offline_deploy_service.yaml
elif [ $scope == "full" ]
then
  ansible-playbook -vv set_global_env.yaml
  ansible-playbook -vv offline_install_packages.yaml
  ansible-playbook -vv offline_load_images.yaml
  ansible-playbook -vv init_kubernetes.yaml
  ansible-playbook -vv clean_services.yaml
  ansible-playbook -vv offline_deploy_service.yaml
else
  echo "Wrong deploy scope variable defined."
fi