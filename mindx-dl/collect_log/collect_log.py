# coding: utf-8

#  Copyright (C)  2020. Huawei Technologies Co., Ltd. All rights reserved.
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

import os
import platform
import stat
import socket
import shutil
import tarfile
import time
import gzip


def get_create_time():
    return "%04d_%02d_%02d_%02d_%02d_%02d" % time.localtime()[:6]


def compress_and_copy_files(src, dst):
    """compress if the src file is uncompressed else copied"""
    print("compress files:" + src)
    if src.lower().endswith(".gz"):
        try:
            shutil.copy(src, dst)
        except OSError as reason:
            print("unable to copy file. %s" % reason)
    else:
        dst = dst + ".gz"
        with gzip.open(dst, 'wb') as f_open_compress, \
                open(src, 'rb') as f_open_src:
            f_open_compress.writelines(f_open_src)


def compress_mindx_files(dst_src_paths):
    for dst_path, _ in dst_src_paths:
        if not os.path.exists(dst_path):
            os.makedirs(dst_path)
    for dst_path, src_path in dst_src_paths:
        if not os.path.exists(src_path):
            print("warning: %s not exists" % src_path)
            continue

        get_compress_file_by_path(dst_path, src_path)


def get_compress_file_by_path(dst_path, src_path):
    file_names = os.listdir(src_path)
    for tmp_file in file_names:
        src = os.path.join(src_path, tmp_file)
        if os.path.isfile(src):
            dst = os.path.join(dst_path, tmp_file)
            try:
                compress_and_copy_files(src, dst)
            except OSError as reason:
                print("error: %s, skipping: %s\n" % (src, reason))


def compress_os_files(base):
    sys_str = platform.platform().lower()

    if "ubuntu" in sys_str:
        os_log_file = "syslog"
    elif "centos" in sys_str:
        os_log_file = "message"
    elif "debian" in sys_str:
        os_log_file = "syslog"
    else:
        print("not support os information: %s\n" % sys_str)
        return

    log_path = []
    os_log_path = "/var/log/"
    files = os.listdir(os_log_path)
    for tmp_file in files:
        if os_log_file in tmp_file:
            log_file_path = os.path.join(os_log_path, tmp_file)
            log_path.append(log_file_path)

    dst_path = os.path.join(base, "os_log")
    os.mkdir(dst_path)
    for tmp_file in log_path:
        if os.path.isfile(tmp_file):
            dst = os.path.join(dst_path, os.path.split(tmp_file)[1])
            compress_and_copy_files(tmp_file, dst)


def get_mindx_dl_compress_files(base, dst_src_file_list):
    compress_mindx_files(dst_src_file_list)
    compress_os_files(base)


def set_log_report_file_path():
    time_base = get_create_time()
    host_name = socket.gethostname()
    tmp_path = "MindX_Report_" + time_base
    base = os.path.join(tmp_path, "LogCollect")
    tar_file_path = "-".join([tmp_path, host_name, "LogCollect.tar.gz"])

    # create folders
    print("creating dst folder:" + base)
    os.makedirs(base)

    return base, tmp_path, tar_file_path


def get_log_path_src_and_dst(base):
    # compress all files from source folders into destination folders
    dst_src_paths = \
        [(os.path.join(base, "volcano-scheduler"),
          "/var/log/atlas_dls/volcano-scheduler"),
         (os.path.join(base, "volcano-admission"),
          "/var/log/atlas_dls/volcano-admission"),
         (os.path.join(base, "volcano-controller"),
          "/var/log/atlas_dls/volcano-controller"),
         (os.path.join(base, "hccl-controller"),
          "/var/log/atlas_dls/hccl-controller"),
         (os.path.join(base, "devicePlugin"), "/var/log/devicePlugin"),
         (os.path.join(base, "cadvisor"), "/var/log/cadvisor"),
         (os.path.join(base, "npuSlog"), "/var/log/npu/slog/host-0/"),
         (os.path.join(base, "apigw"), "/var/log/atlas_dls/apigw"),
         (os.path.join(base, "cec"), "/var/log/atlas_dls/cec"),
         (os.path.join(base, "dms"), "/var/log/atlas_dls/dms"),
         (os.path.join(base, "mms"), "/var/log/atlas_dls/mms"),
         (os.path.join(base, "mysql"), "/var/log/atlas_dls/mysql"),
         (os.path.join(base, "nginx"), "/var/log/atlas_dls/nginx"),
         (os.path.join(base, "tjm"), "/var/log/atlas_dls/tjm")]

    return dst_src_paths


def create_compress_file(tmp_path, tar_file_path):
    # create a tar file, and archive all compressed files into ita
    print("create tar file:" + tar_file_path + ", from all compressed files")
    try:
        os.path.abspath(tar_file_path)
        with tarfile.open(tar_file_path, 'w:gz') as tmp_file:
            tmp_file.add(tmp_path)
            print("adding to tar: %s\n" % tmp_path)
    except tarfile.TarError as err:
        print("error: %s, skipping: %s\n" % (tmp_path, err))


def set_file_right(tar_file_path):
    os.chmod(tar_file_path, stat.S_IREAD)


def delete_tmp_file(tmp_path):
    print("delete temp folder" + tmp_path)
    shutil.rmtree(tmp_path)


def main():
    print("begin to collect log files")

    base, tmp_path, tar_file_path = set_log_report_file_path()

    dst_src_paths = get_log_path_src_and_dst(base)

    get_mindx_dl_compress_files(base, dst_src_paths)

    create_compress_file(base, tar_file_path)

    set_file_right(tar_file_path)

    delete_tmp_file(tmp_path)

    print("collect log files finish")


if __name__ == '__main__':
    main()
