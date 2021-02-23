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

import yaml
from jinja2 import Template
from kubernetes import client
from kubernetes import config


def get_core_v1_api():
    """
    :return: api client
    """
    config.load_kube_config()
    core_api = client.CoreV1Api()

    return core_api


def get_custom_obj_api():
    """
    :return: custom api client
    """
    config.load_kube_config()
    custom_api = client.CustomObjectsApi()

    return custom_api


def get_batch_v1_api():
    """
    :return: batch api client
    """
    config.load_kube_config()
    batch_v1_api = client.BatchV1Api()

    return batch_v1_api


def get_app_v1_api():
    """
    :return: app api client
    """
    config.load_kube_config()
    app_api = client.AppsV1Api()

    return app_api


def render_template(template_file, **kwargs):
    """render variables in the yaml file."""
    try:
        with open(template_file, "r", encoding='utf-8') as f_open:
            template_content = f_open.read()
    except FileNotFoundError:
        template_content = ''

    if not template_content:
        return {}

    template = Template(template_content)
    yaml_string = template.render(**kwargs)
    yaml_json = yaml.safe_load(yaml_string)

    return yaml_json


def get_pod_list_by_namespace(namespace):
    """
    :return: the JSON information about pods in a specified namespace.
    """
    core_api = get_core_v1_api()
    res_list = core_api.list_namespaced_pod(namespace).items
    pods_info_json = [item.to_dict() for item in res_list]

    return pods_info_json


def get_vcjob_log(pod_name, namespace):
    """log of vcjob"""
    core_api = get_core_v1_api()
    result = core_api.read_namespaced_pod_log(namespace=namespace,
                                              name=pod_name)

    print("=====get vcjob logs: {}".format(result))
