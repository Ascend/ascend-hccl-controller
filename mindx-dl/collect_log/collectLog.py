#!/usr/bin/python

# ------------------------------------------------------------------------------
#   Copyright (C), 2020, Huawei Tech. Co., Ltd.
# ------------------------------------------------------------------------------
# Filename: collectLog.py
# Requires: python 3.7 (or newer version)
# Desc:     Gather and compress Mindx-DL logs into a single tar files.
#
# ------------------------------------------------------------------------------

import gzip
import os
import platform
import socket
import sys
import tarfile
import time
import shutil
from pwd import getpwnam


def log(content):
    print(content)


def get_create_time():
    return "MindX-DL_Report_%04d_%02d_%02d_%02d_%02d_%02d" % time.localtime()[:6]


def compress(src, dst):
    """compress if the src file is uncompressed"""
    log("compress files:" + src)
    if src[-3:].lower() == ".gz":
        f = open
    else:
        f = gzip.open
        dst = dst + ".gz"
    fOpenDst = f(dst, 'wb', 9)
    fOpenSrc = open(src, 'rb')
    fOpenDst.writelines(fOpenSrc)
    fOpenDst.close()
    fOpenSrc.close()
    return dst


def get_compress_file_paths(tmp_path, dst_src_paths, done):
    for dst_path, src_path in dst_src_paths:
        if not os.path.exists(dst_path):
            os.makedirs(dst_path)
    for dst_path, src_path in dst_src_paths:
        filenames = []
        for _, _, filenames in os.walk(src_path):
            break
        for file in filenames:
            src = os.path.join(src_path, file)
            dst = os.path.join(dst_path, file)
            log("Compressing: %s\n" % dst)
            try:
                dst = compress(src, dst)
                done.append(dst[len(tmp_path) + 1:])
            except OSError as reason:
                log("error:%s, skipping: %s\n" % (src, reason))
    return done


def compress_os_files(base, tmp_path, done):
    sysStr = platform.platform().lower()

    if "ubuntu" in sysStr:
        os_log_path = "/var/log/syslog*"
    elif "centos" in sysStr:
        os_log_path = "/var/log/message*"
    else:
        os_log_path = ""

    filepath = os.popen("ls " + os_log_path)
    log_path = filepath.read()
    log_path = log_path.strip('\n').replace("\n", " ").split(" ")
    for dst_path, fileList in [(base + "/OS_log", log_path)]:
        os.mkdir(dst_path)
        for file in fileList:
            if os.path.isfile(file):
                dst = os.path.join(dst_path, os.path.split(file)[1])
                sys.stdout.write("Compressing: %s\n" % dst)
                dst = compress(file, dst)
                done.append(dst[len(tmp_path) + 1:])
    return done


def compress_file_list(base, tmp_path, dst_src_file_list, done):
    for dst_path, file_list in dst_src_file_list:
        for file in file_list:
            if os.path.isfile(file):
                dst = os.path.join(dst_path, os.path.split(file)[1])
                log("Compressing: %s\n" % dst)
                dst = compress(file, dst)
                done.append(dst[len(tmp_path) + 1:])

    done = compress_os_files(base, tmp_path, done)
    return done


def get_mindx_dl_compress_files(base, tmp_path, dst_src_file_list, done):
    done = get_compress_file_paths(tmp_path, dst_src_file_list, done)
    done = compress_file_list(base, tmp_path, dst_src_file_list, done)
    return done


def set_log_report_file_path():
    time_base = get_create_time()
    hostName = socket.gethostname()
    tmp_path = "/tmp/MindXReport/" + time_base
    base = os.path.join(tmp_path, "LogCollect")
    tarFilePath = tmp_path + "-" + hostName + "-LogCollect.zip"

    # create folders
    log("Creating dst folder:" + base)
    os.makedirs(tmp_path)
    os.makedirs(base)

    return base, tmp_path, tarFilePath


def get_log_path_src_and_dst(base):
    # compress all files from source folders into destination folders
    dst_src_paths = [(base + "/volcano-scheduler", "/var/log/atlas_dls/volcano-scheduler"),
                     (base + "/volcano-admission", "/var/log/atlas_dls/volcano-admission"),
                     (base + "/volcano-controller", "/var/log/atlas_dls/volcano-controller"),
                     (base + "/hccl-controller", "/var/log/atlas_dls/hccl-controller"),
                     (base + "/devicePlugin", "/var/log/devicePlugin"),
                     (base + "/cadvisor", "/var/log/cadvisor"),
                     (base + "/npuSlog", "/var/log/npu/slog/host-0/"),
                     (base + "/apigw", "/var/log/npu/slog/host-0/"),
                     (base + "/cec", "/var/log/npu/slog/host-0/"),
                     (base + "/dms", "/var/log/npu/slog/host-0/"),
                     (base + "/mms", "/var/log/npu/slog/host-0/"),
                     (base + "/mysql", "/var/log/npu/slog/host-0/"),
                     (base + "/nginx", "/var/log/npu/slog/host-0/"),
                     (base + "/tjm", "/var/log/npu/slog/host-0/")]

    return dst_src_paths


def create_compress_file(done, tmp_path, tarFilePath):
    # create a tar file, and archive all compressed files into ita
    log("create a tar file:" + tarFilePath + ", and archive all compressed files into it")
    tar = tarfile.open(tarFilePath, 'w')
    old_path = os.getcwd()
    os.chdir(tmp_path)
    for filename in done:
        try:
            tar.add(filename)
            log("Adding to tar: %s\n" % filename)
        except OSError as reason:
            log("error: %s, skipping: %s\n" % (filename, reason))
    tar.close()
    os.chdir(old_path)


def set_file_right(tarFilePath):
    uid = getpwnam("hwMindX").pw_uid
    os.lchown(tarFilePath, uid, uid)


def delete_tmp_file(tmp_path):
    log("Delete temp folder" + tmp_path)
    shutil.rmtree(tmp_path)


def main():
    log("begin to collect log files")

    base, tmp_path, tarFilePath = set_log_report_file_path()

    dst_src_paths = get_log_path_src_and_dst(base)

    done = []
    done = get_mindx_dl_compress_files(base, tmp_path, dst_src_paths, done)

    create_compress_file(done, tmp_path, tarFilePath)

    set_file_right(tarFilePath)

    delete_tmp_file(tmp_path)

    log("collect log files finish")


if __name__ == '__main__':
    main()
