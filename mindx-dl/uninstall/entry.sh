#!/bin/bash
# Entry script for offline cluster deployment.
# Copyright © Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e
ansible-playbook -vv uninstall.yaml