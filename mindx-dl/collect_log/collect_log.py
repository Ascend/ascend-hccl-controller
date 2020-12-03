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
import stat
import socket
import shutil
import tarfile
import time
import logging


def init_sys_log():
    logging.basicConfig(level=logging.INFO)


def get_create_time():
    return "MindX_Report_%04d_%02d_%02d_%02d_%02d_%02d" % time.localtime()[:6]


def compress(src, dst):
    """compress if the src file is uncompressed"""
    logging.info("compress files:" + src)
    if src.lower().endswith(".gz"):
        f_open = open
    else:
        f_open = gzip.open
        dst = dst + ".gz"
    with f_open(dst, 'wb', 9) as f_open_dst, open(src, 'rb') as f_open_src:
        f_open_dst.writelines(f_open_src)

    return dst


def get_compress_file_paths(tmp_path, dst_src_paths, done):
    for dst_path, src_path in dst_src_paths:
        if not os.path.exists(dst_path):
            os.makedirs(dst_path)
    for dst_path, src_path in dst_src_paths:
        if not os.path.exists(src_path):
            logging.warning("%s not exists" % src_path)
            continue

        file_names = []
        for _, _, file_names in os.walk(src_path):
            break
        for tmp_file in file_names:
            src = os.path.join(src_path, tmp_file)
            dst = os.path.join(dst_path, tmp_file)
            logging.info("Compressing: %s\n" % dst)
            try:
                dst = compress(src, dst)
                done.append(dst[len(tmp_path) + 1:])
            except OSError as reason:
                logging.error("error:%s, skipping: %s\n" % (src, reason))
    return done


def compress_os_files(base, tmp_path, done):
    sys_str = platform.platform().lower()

    if "ubuntu" in sys_str:
        os_log_file = "syslog"
    elif "centos" in sys_str:
        os_log_file = "message"
    elif "debian" in sys_str:
        os_log_file = "syslog"
    else:
        logging.error("not support os inf %s\n" % sys_str)
        return done

    log_path = []
    os_log_path = "/var/log/"
    files = os.listdir(os_log_path)
    for tmp_file in files:
        if os_log_file in tmp_file:
            log_file_path = os_log_path + tmp_file
            log_path.append(log_file_path)

    dst_path = base + "/OS_log"
    os.mkdir(dst_path)
    for tmp_file in log_path:
        if os.path.isfile(tmp_file):
            dst = os.path.join(dst_path, os.path.split(tmp_file)[1])
            dst = compress(tmp_file, dst)
            relative_path = dst[len(tmp_path) + 1:]
            done.append(relative_path)
    return done


def compress_file_list(base, tmp_path, dst_src_file_list, done):
    for dst_path, file_list in dst_src_file_list:
        for tmp_file in file_list:
            if os.path.isfile(tmp_file):
                dst = os.path.join(dst_path, os.path.split(tmp_file)[1])
                logging.info("Compressing: %s\n" % dst)
                dst = compress(tmp_file, dst)
                done.append(dst[len(tmp_path) + 1:])

    done = compress_os_files(base, tmp_path, done)
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
    tar_file_path = tmp_path + "-" + host_name + "-LogCollect.gz"

    # create folders
    logging.info("Creating dst folder:" + base)
    os.makedirs(tmp_path)
    os.makedirs(base)

    return base, tmp_path, tar_file_path


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
    logging.info("create tar file:" + tar_file_path + ", from all compressed files")
    try:
        with tarfile.open(tar_file_path, 'w:gz') as tmp_file:
            old_path = os.getcwd()
            os.chdir(tmp_path)
            for filename in done:
                tmp_file.add(filename)
                logging.info("Adding to tar: %s\n" % filename)
            os.chdir(old_path)
    except tarfile.TarError as err:
        logging.error("error: %s, skipping: %s\n" % (filename, err))


def set_file_right(tar_file_path):
    os.chmod(tar_file_path, stat.S_IREAD)


def delete_tmp_file(tmp_path):
    logging.info("Delete temp folder" + tmp_path)
    shutil.rmtree(tmp_path)


def main():
    init_sys_log()

    logging.info("begin to collect log files")

    base, tmp_path, tar_file_path = set_log_report_file_path()

    dst_src_paths = get_log_path_src_and_dst(base)

    done = []
    done = get_mindx_dl_compress_files(base, tmp_path, dst_src_paths, done)

    create_compress_file(done, tmp_path, tar_file_path)

    set_file_right(tar_file_path)

    delete_tmp_file(tmp_path)

    logging.info("collect log files finish")


if __name__ == '__main__':
    main()
