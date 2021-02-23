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

from configmap_crud import create_config_map
from configmap_crud import delete_config_map
from configmap_crud import get_config_map_param_dict
from utils import get_core_v1_api
from utils import get_custom_obj_api
from utils import get_vcjob_log
from utils import render_template
from vcjob_config_sample import JOB_NAME
from vcjob_config_sample import JOB_PARAMS_1P

VCJOB_GROUP = 'batch.volcano.sh'
VCJOB_API_VERSION = 'v1alpha1'
VCJOB_PLURAL = 'jobs'


def create_job(api_obj, **vcjob_config_dict):
    """create vcjob"""
    result = api_obj.create_namespaced_custom_object(
        group=VCJOB_GROUP,
        version=VCJOB_API_VERSION,
        namespace='default',
        plural=VCJOB_PLURAL,
        body=vcjob_config_dict)

    print("=====create vcjob: {}".format(result))


def delete_job(api_obj):
    """delete vcjob"""
    result = api_obj.delete_namespaced_custom_object(
        group=VCJOB_GROUP,
        version=VCJOB_API_VERSION,
        namespace='default',
        plural=VCJOB_PLURAL,
        name=JOB_NAME)

    print("=====delete vcjob: {}".format(result))


def get_single_vcjob_info(api_obj):
    """get single vcjob"""
    result = api_obj.get_namespaced_custom_object(
        group=VCJOB_GROUP,
        version=VCJOB_API_VERSION,
        namespace='default',
        plural=VCJOB_PLURAL,
        name=JOB_NAME)

    print("=====get single vcjob: {}".format(result))


def get_namespace_vcjob(api_job):
    """get vcjob by namespace"""
    result = api_job.list_namespaced_custom_object(
        group=VCJOB_GROUP,
        version=VCJOB_API_VERSION,
        namespace='default',
        plural=VCJOB_PLURAL)

    print("=====get namespaced vcjob: {}".format(result))


def get_vcjob_param_dict():
    """render vcjob config file"""
    job_desc_file = "vcjob.yaml"
    vcjob_config_dict = render_template(job_desc_file, **JOB_PARAMS_1P)

    return vcjob_config_dict


def create_vcjob(custom_v1_api, core_v1_api):
    """entry for creating vcjob"""
    cfg_map_dict = get_config_map_param_dict()
    vcjob_config_dict = get_vcjob_param_dict()

    # don't change the order
    create_config_map(core_v1_api, **cfg_map_dict)
    create_job(custom_v1_api, **vcjob_config_dict)


def delete_vcjob(custom_v1_api, core_v1_api):
    """entry for deleting vcjob"""
    delete_job(custom_v1_api)
    delete_config_map(core_v1_api)


def main():
    # get api client
    custom_api = get_custom_obj_api()
    core_api = get_core_v1_api()

    create_vcjob(custom_api, core_api)

    get_namespace_vcjob(custom_api)

    get_single_vcjob_info(custom_api)

    # in this case, podname is 'test-job-default-test-0'
    pod_name = 'test-job-default-test-0'
    get_vcjob_log(pod_name, 'default')

    delete_vcjob(custom_api, core_api)


if __name__ == '__main__':
    main()
