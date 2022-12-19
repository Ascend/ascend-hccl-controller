#!/usr/bin/env bash
set -o errexit

# 1. cephfs：

# 事前准备：创建cephfs的挂载目录，并手动挂载cephfs存储集群到该目录。<CEPHFS_USER>默认为admin，<CEPHFS_KEY>可在cephfs monitor节点通过`ceph auth get-key client.admin`查询

# mkdir <cephfs的挂载目录>
# mount -t ceph <CEPHFS_IP>:<CEPHFS_PORT>:/ <cephfs的挂载目录> -o name=<CEPHFS_USER>,secret=<CEPHFS_KEY>
# cd <cephfs的挂载目录>

# 将本文件create_storage_dir.sh和本工具目录的playbooks/roles/mindx.nfs.server/files/rule.yaml文件拷贝到<cephfs的挂载目录>
# 在<cephfs的挂载目录>下，执行`bash create_storage_dir.sh`，即可创建STORAGE_PATH目录及其下的相关目录（默认为"data/atlas_dls"，根据group_vars/all.yaml里STORAGE_PATH实际配置更改，注意不是"/data/atlas_dls"）



# 2. oceanstore：

# 由于使用hostpath方式使用oceanstore，需要在所有k8s节点上执行，并建议直接按如下操作创建oceanstore的挂载目录
# 事前准备：已安装好oceanstore的dpc客户端（由oceanstore存储完成）；创建oceanstore的挂载目录，并手动挂载oceanstore存储集群到该目录。

# mkdir /dl  # 创建oceanstore的挂载目录
# chown 9000:9000 /dl  # 使用hostpath方式，需要将挂载目录属主设置为9000
# mount -t dpc <oceanstore标识符> /dl  # <oceanstore标识符>，由oceanstore存储提供，不可与“/dl”同名
# 使用autofs设置dpc开机自动挂载，以达到高可用  # 具体操作由oceanstore存储提供

# 以上步骤在各个k8s节点上执行完毕后，如下操作只需要在任一节点执行即可
# 将本文件create_storage_dir.sh和本工具目录的playbooks/roles/mindx.nfs.server/files/rule.yaml文件拷贝到根目录/
# group_vars/all.yaml里STORAGE_PATH修改为/dl/atlas_dls；如下STORAGE_PATH变量修改为dl/atlas_dls
# 在根目录/下，执行`bash create_storage_dir.sh`



# 根据实际情况修改STORAGE_PATH变量
STORAGE_PATH=data/atlas_dls

if [[ ! -f rule.yaml ]]; then
    echo "rule.yaml is not existed, please copy here from playbooks/roles/mindx.nfs.server/files/rule.yaml"
    exit
fi

mkdir -p -m 750 ${STORAGE_PATH} ${STORAGE_PATH}/platform
cd ${STORAGE_PATH}/platform
mkdir -p -m 750 kmc log loki services services/prometheus
cd log
mkdir -p -m 750 apigw apigw-business cluster-manager data-manager dataset-manager image-manager model-manager inference-manager train-manager user-manager alarm-manager
cd ../../../..
cp -f rule.yaml ${STORAGE_PATH}/platform/services/prometheus/rule.yaml
chmod 600 ${STORAGE_PATH}/platform/services/prometheus/rule.yaml
chown -R 9000:9000 ${STORAGE_PATH}
chown -R root:root ${STORAGE_PATH}/platform/log/image-manager
