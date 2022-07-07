#!/usr/bin/env bash
set -o errexit

# 事前准备：创建cephfs的挂载目录，并手动挂载cephfs存储集群到该目录。<CEPHFS_USER>一般为admin，<CEPHFS_KEY>可在cephfs monitor节点通过`ceph auth get-key client.admin`查询

# mkdir <cephfs的挂载目录>
# mount -t ceph <CEPHFS_IP>:<CEPHFS_PORT>:/ <cephfs的挂载目录> -o name=<CEPHFS_USER>,secret=<CEPHFS_KEY>
# cd <cephfs的挂载目录>

# 将本文件create_ceph_dir.sh和本工具目录的playbooks/roles/mindx.nfs.server/files/rule.yaml文件拷贝到<cephfs的挂载目录>
# 在<cephfs的挂载目录>下，执行`bash create_ceph_dir.sh`，即可创建STORAGE_PATH目录及其下的相关目录（默认为"data/atlas_dls"，根据group_vars/all.yaml里STORAGE_PATH实际配置更改，注意不是"/data/atlas_dls"）

if [[ ! -f rule.yaml ]]; then
    echo "rule.yaml is not existed, please copy here from playbooks/roles/mindx.nfs.server/files/rule.yaml"
    exit
fi
STORAGE_PATH=data/atlas_dls
mkdir -p -m 750 ${STORAGE_PATH} ${STORAGE_PATH}/platform
cd ${STORAGE_PATH}/platform
mkdir -p -m 750 kmc log services services/prometheus
cd log
mkdir -p -m 750 apigw cluster-manager data-manager dataset-manager edge-manager image-manager label-manager model-manager task-manager train-manager user-manager alarm-manager
cd ../../../..
cp -f rule.yaml ${STORAGE_PATH}/platform/services/prometheus/rule.yaml
chmod 600 ${STORAGE_PATH}/platform/services/prometheus/rule.yaml
chown -R 9000:9000 ${STORAGE_PATH}
chown -R root:root ${STORAGE_PATH}/platform/log/image-manager
