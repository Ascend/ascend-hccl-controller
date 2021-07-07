# coding: UTF-8

#  Copyright (C)  2021. Huawei Technologies Co., Ltd. All rights reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

"""
This file is a parameter sample for delivering A800-9000/9010 and 300T
training tasks. Please modify the value of each variable according to
the actual situation.
"""

JOB_NAME = 'test-job'

# Atlas 800-9000/Atlas 800-9010
# 1p
JOB_PARAMS_1P = {
    'job_name': JOB_NAME,
    'npu_count': 1,
    'node_count': 1,
    'max_retry': 3,
    'docker_image': 'ubuntu:18.04',
    'command': [
        '/bin/bash', '-c',
        'cd /job/code/ModelZoo_Resnet50_HC;bash train_start.sh'
    ],
    'args': [],
    'node_selector': {
        'host-arch': 'huawei-x86'  # huawei-x86/huawei-arm
    },
    'container_code_path': '/job/code',
    'container_data_path': '/job/data',
    'container_output_path': '/job/output',
    'use_nfs': True,
    # if use_nfs is False, 'nfs_server_ip' should be set to ''
    # if use_nfs is True, 'nfs_server_ip' should be set to NFS server address
    'nfs_server_ip': '127.0.0.1',
    'host_code_path': '/data/mindx-dl/code',
    'host_data_path': '/data/mindx-dl/public/',
    'host_output_path': '/data/mindx-dl/output'
}

# Atlas 800-9000/Atlas 800-9010
# 8p
JOB_PARAMS_8P = JOB_PARAMS_1P.copy()
JOB_PARAMS_8P.update(**{'npu_count': 8})

# Atlas 800-9000/Atlas 800-9010
# 16p
JOB_PARAMS_16P = JOB_PARAMS_1P.copy()
JOB_PARAMS_16P.update(**{'npu_count': 8, 'node_count': 2})

# Atlas 300T
# 1p
JOB_PARAMS_300T_1P = JOB_PARAMS_1P.copy()
JOB_PARAMS_300T_1P.update(**{'node_selector': {'host-arch': 'huawei-x86',
                                               'accelerator-type': 'card'}})

# Atlas 300T
# 4p
JOB_PARAMS_300T_4P = JOB_PARAMS_1P.copy()
JOB_PARAMS_300T_4P.update(**{'node_selector': {'host-arch': 'huawei-x86',
                                               'accelerator-type': 'card'},
                             'npu_count': 2,
                             'node_count': 2})
