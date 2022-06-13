#!/usr/bin/env bash
set -o errexit

# 事前准备：创建cephfs的挂载目录，并手动挂载cephfs存储集群到该目录。<CEPHFS_KEY_BASE64>可在cephfs monitor节点通过`ceph auth get-key client.admin | base64`查询

# mkdir <cephfs的挂载目录>
# mount -t ceph <CEPHFS_IP>:<CEPHFS_PORT>:/ <cephfs的挂载目录> -o name=<CEPHFS_USER>,secret=<CEPHFS_KEY_BASE64>
# cd <cephfs的挂载目录>

# 将本脚本拷贝到挂载目录目录下执行，即创建STORAGE_PATH目录及其下的相关目录（默认为"data/atlas_dls"，根据group_vars/all.yaml里STORAGE_PATH实际配置更改，注意不是"/data/atlas_dls"）
STORAGE_PATH=data/atlas_dls
mkdir -p -m 750 ${STORAGE_PATH} ${STORAGE_PATH}/platform
cd ${STORAGE_PATH}/platform
mkdir -p -m 750 kmc log
cd log
mkdir -p -m 750 apigw cluster-manager data-manager dataset-manager edge-manager image-manager label-manager model-manager task-manager train-manager user-manager alarm-manager
cd ../../../..
chown -R 9000:9000 ${STORAGE_PATH}
chown -R root:root ${STORAGE_PATH}/platform/log/image-manager
