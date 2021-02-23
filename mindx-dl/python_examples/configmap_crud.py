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

from utils import get_core_v1_api
from utils import render_template
from vcjob_config_sample import JOB_NAME


def create_configmap(api_obj, **cfg_map_dict):
    """create configmap"""
    result = api_obj.create_namespaced_config_map(
        namespace='default', body=cfg_map_dict)

    print("=====create configmap: {}".format(result))


def get_configmap(api_obj):
    """get configmap"""
    result = api_obj.list_namespaced_config_map(
        namespace='default')

    print("=====get configmap: {}".format(result))


def delete_configmap(api_obj):
    """delete configmap"""
    cfg_name = 'rings-config-' + JOB_NAME
    result = api_obj.delete_namespaced_config_map(
        name=cfg_name, namespace='default')

    print("=====delete configmap: {}".format(result))


def main():
    core_v1_api = get_core_v1_api()

    # create configmap
    cfg_map_file = "vcjob_configmap.yaml"
    cfg_map_params = {'job_name': JOB_NAME}
    cfg_map_dict = render_template(cfg_map_file, **cfg_map_params)
    create_configmap(core_v1_api, **cfg_map_dict)

    get_configmap(core_v1_api)

    delete_configmap(core_v1_api)


if __name__ == '__main__':
    main()
