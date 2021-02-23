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

from configmap_crud import create_configmap
from configmap_crud import delete_configmap
from utils import get_app_v1_api
from utils import get_core_v1_api
from utils import get_pods_list_by_namespace
from utils import render_template
from vcjob_config_sample import JOB_NAME
from vcjob_config_sample import JOB_PARAMS_1P


def create_deployment(api_obj, **vcjob_deployment_dict):
    """create a deployment vcjob"""
    result = api_obj.create_namespaced_deployment(
        namespace='default', body=vcjob_deployment_dict)

    print('=====create deployment vcjob: {}'.format(result))


def delete_deployment(api_job):
    """delete deployment vcjob"""
    result = api_job.delete_namespaced_deployment(
        name=JOB_NAME,
        namespace='default',
        body=client.V1DeleteOptions(propagation_policy='Foreground',
                                    grace_period_seconds=5))

    print('=====delete deployment vcjob: {}'.format(result))


def get_pods_info_by_namespace():
    """query deployment vcjob"""
    result = get_pods_list_by_namespace('default')

    print('=====query deployment vcjob: {}'.format(result))


def create_deployment_vcjob(app_api, core_api):
    """entry for creating vcjob"""
    cfg_map_file = "vcjob_configmap.yaml"
    cfg_map_params = {'job_name': JOB_NAME}
    cfg_map_dict = render_template(cfg_map_file, **cfg_map_params)

    deployment_file = "vcjob_deployment.yaml"
    vcjob_deployment_dict = render_template(deployment_file, **JOB_PARAMS_1P)

    # don't change the order
    create_configmap(core_api, **cfg_map_dict)
    create_deployment(app_api, **vcjob_deployment_dict)


def delete_deployment_vcjob(app_api, core_api):
    """entry for deleting vcjob"""
    delete_deployment(app_api)
    delete_configmap(core_api)


def main():
    # get api client
    app_v1_api = get_app_v1_api()
    core_api = get_core_v1_api()

    create_deployment_vcjob(app_v1_api, core_api)

    get_pods_info_by_namespace()

    delete_deployment_vcjob(app_v1_api, core_api)


if __name__ == '__main__':
    main()
