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
create, delete, query
"""

from kubernetes import client

from utils import get_batch_v1_api
from utils import get_pod_list_by_namespace
from utils import render_template

INFER_JOB_NAME = 'infer-test'


def get_infer_param_dict():
    """render infer config file"""
    infer_config_params = {
        'infer_job_name': INFER_JOB_NAME,
        'node_selector': {
            'accelerator': 'huawei-Ascend310',
            'host-arch': 'huawei-arm'  # huawei-arm/huawei-x86
        },
        'npu_count': 1,
        'docker_image': 'ubuntu:18.04'
    }
    infer_yaml_file = 'infer.yaml'
    # render the infer.yaml file
    infer_config_dict = render_template(infer_yaml_file,
                                        **infer_config_params)

    return infer_config_dict


def create_infer_job(api_obj):
    """create a inference job"""
    infer_config_dict = get_infer_param_dict()

    result = api_obj.create_namespaced_job(namespace='default',
                                           body=infer_config_dict)

    print("=====create infer job: {}".format(result))


def get_infer_job():
    result_json = get_pod_list_by_namespace('default')

    print("=====query infer job in namespace: {}".format(result_json))


def delete_infer_job(api_obj):
    """delete infer job"""
    result = api_obj.delete_namespaced_job(
        name=INFER_JOB_NAME,
        namespace='default',
        body=client.V1DeleteOptions(propagation_policy='Foreground',
                                    grace_period_seconds=5))

    print("=====delete infer job: {}".format(result))


def main():
    # get api client
    batch_api = get_batch_v1_api()

    create_infer_job(batch_api)

    get_infer_job()

    delete_infer_job(batch_api)


if __name__ == '__main__':
    main()
