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

import gzip
import os
import platform
import socket
import tarfile
import time
import shutil

from pwd import getpwnam


def log(content):
    print(content)


def get_create_time():
    return "MindX_Report_%04d_%02d_%02d_%02d_%02d_%02d" % time.localtime()[:6]


def compress(src, dst):
    """compress if the src file is uncompressed"""
    log("compress files:" + src)
    if src[-3:].lower() == ".gz":
        f_open = open
    else:
        f_open = gzip.open
        dst = dst + ".gz"
    f_open_dst = f_open(dst, 'wb', 9)
    f_open_src = open(src, 'rb')
    f_open_dst.writelines(f_open_src)
    f_open_dst.close()
    f_open_src.close()
    return dst


def get_compress_file_paths(tmp_path, dst_src_paths, done):
    for dst_path, src_path in dst_src_paths:
        if not os.path.exists(dst_path):
            os.makedirs(dst_path)
    for dst_path, src_path in dst_src_paths:
        file_names = []
        for _, _, file_names in os.walk(src_path):
            break
        done = compress_files(done, dst_path, file_names, tmp_path)

    return done


def compress_os_files(base, tmp_path, done):
    sys_str = platform.platform().lower()

    if "ubuntu" in sys_str:
        os_log_path = "/var/log/syslog*"
    elif "centos" in sys_str:
        os_log_path = "/var/log/message*"
    else:
        os_log_path = ""

    log_path = ""
    try:
        filepath = os.listdir(os_log_path)
        log_path = filepath.read()
    except OSError as ex:
        log(ex)

    log_path = log_path.strip('\n').replace("\n", " ").split(" ")
    for dst_path, file_list in [(base + "/OS_log", log_path)]:
        os.mkdir(dst_path)
        done = compress_files(done, dst_path, file_list, tmp_path)

    return done


def compress_file_list(base, tmp_path, dst_src_file_list, done):
    for dst_path, file_list in dst_src_file_list:
        done = compress_files(done, dst_path, file_list, tmp_path)

    done = compress_os_files(base, tmp_path, done)
    return done


def compress_files(done, dst_path, file_list, tmp_path):
    for file in file_list:
        if os.path.isfile(file):
            dst_file = os.path.join(dst_path, os.path.split(file)[1])
            log("Compressing: %s\n" % dst_file)
            try:
                dst_file = compress(file, dst_file)
                done.append(dst_file[len(tmp_path) + 1:])
            except OSError as reason:
                log("error:%s, skipping: %s\n" % (file, reason))

    return done


def get_mindx_dl_compress_files(base, tmp_path, dst_src_file_list, done):
    done = get_compress_file_paths(tmp_path, dst_src_file_list, done)
    done = compress_file_list(base, tmp_path, dst_src_file_list, done)
    return done


def set_log_report_file_path():
    time_base = get_create_time()
    host_name = socket.gethostname()
    tmp_path = os.path.join(os.getcwd(), time_base)
    base = os.path.join(tmp_path, "LogCollect")
    tar_file_ath = tmp_path + "-" + host_name + "-LogCollect.gz"

    # create folders
    log("Creating dst folder:" + base)
    os.makedirs(tmp_path)
    os.makedirs(base)

    return base, tmp_path, tar_file_ath


def get_log_path_src_and_dst(base):
    # compress all files from source folders into destination folders
    dst_src_paths = \
        [(base + "/volcano-scheduler",
          "/var/log/atlas_dls/volcano-scheduler"),
         (base + "/volcano-admission",
          "/var/log/atlas_dls/volcano-admission"),
         (base + "/volcano-controller",
          "/var/log/atlas_dls/volcano-controller"),
         (base + "/hccl-controller",
          "/var/log/atlas_dls/hccl-controller"),
         (base + "/devicePlugin", "/var/log/devicePlugin"),
         (base + "/cadvisor", "/var/log/cadvisor"),
         (base + "/npuSlog", "/var/log/npu/slog/host-0/"),
         (base + "/apigw", "/var/log/atlas_dls/apigw"),
         (base + "/cec", "/var/log/atlas_dls/cec"),
         (base + "/dms", "/var/log/atlas_dls/dms"),
         (base + "/mms", "/var/log/atlas_dls/mms"),
         (base + "/mysql", "/var/log/atlas_dls/mysql"),
         (base + "/nginx", "/var/log/atlas_dls/nginx"),
         (base + "/tjm", "/var/log/atlas_dls/tjm")]

    return dst_src_paths


def create_compress_file(done, tmp_path, tar_file_path):
    # create a tar file, and archive all compressed files into ita
    log("create tar file:" + tar_file_path + ", from all compressed files")
    try:
        with tarfile.open(tar_file_path, 'w:gz') as file:
            old_path = os.getcwd()
            os.chdir(tmp_path)
            for filename in done:
                file.add(filename)
                log("Adding to tar: %s\n" % filename)
            os.chdir(old_path)
    except tarfile.TarError as err:
        log("error: %s, skipping: %s\n" % (filename, err))


def set_file_right(tar_file_path):
    uid = getpwnam("hwMindX").pw_uid
    os.lchown(tar_file_path, uid, uid)


def delete_tmp_file(tmp_path):
    log("Delete temp folder" + tmp_path)
    shutil.rmtree(tmp_path)


def main():
    log("begin to collect log files")

    base, tmp_path, tar_file_path = set_log_report_file_path()

    dst_src_paths = get_log_path_src_and_dst(base)

    done = []
    done = get_mindx_dl_compress_files(base, tmp_path, dst_src_paths, done)

    create_compress_file(done, tmp_path, tar_file_path)

    set_file_right(tar_file_path)

    delete_tmp_file(tmp_path)

    log("collect log files finish")


if __name__ == '__main__':
    main()
