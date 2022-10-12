# 快速入门

## 功能描述

本文主要介绍如何使用ansible安装mindxdl平台所需开源软件安装。其中包含如下开源软件

| 软件名      |    版本   | 备注                    |
| ---------- | ----------| ------------------------|
| python3    | 3.6       | ansible会安装到python3，本机节点安装|
| ansible    | 2.11.6    | 任务编排的自动化平台，本机节点安装   |
| docker     | 19.03     | 集群中所有节点都需要安装   |
| harbor     | 2.3.3     | 容器镜像仓                |
| kubernetes | 1.19.16   | k8s平台                  |
| nfs        | 1.3       | nfs存储系统               |
| mysql      | 8.0.26    | 安装在k8s集群中，关系型数据库系统 |
| redis      | 5.0.14    | 安装在k8s集群中，非关系型数据库系统 |
| prometheus + grafana + node-exporter + alertmanager + kube-state-metrics | 2.29.2 + 7.5.5 + 1.2.2 + 0.24.0 + 2.3.0 | 安装在k8s集群中，资源监控组件      |
| chrony     | 3.2        | (可选）时间同步组件，所有节点安装          |

## 环境要求

### 支持的操作系统说明

| 操作系统|    版本   | CPU架构 | 备注         |
|:------:|:---------:|:-------:|:-----------:|
| Ubuntu | 18.04     | aarch64 |安装到【Software selection】这一步时勾选【OpenSSH server】/【SSH server】附件组件|
| Ubuntu | 18.04     | x86_64  |安装到【Software selection】这一步时勾选【OpenSSH server】/【SSH server】附件组件|
| Centos | 7.6       | aarch64 |安装到【SOFTWARE SELECTION】这一步时建议勾选”Debugging Tools、Compatibility Libraries、Development Toos"附件组件|

根目录的磁盘空间利用率高于85%会触发Kubelet的镜像垃圾回收机制，将导致服务不可用。请确保根目录有足够的磁盘空间，建议大于500GB

### 角色说明

1. 管理（master)

master节点无需为NPU插卡环境，普通服务器即可

2. 计算（worker)

| 中心推理硬件    | 中心训练硬件|
|:--------------:|:----------:|
| A300-3000      | A800-9000  |
| A300-3010      | A800-9010  |
| Atlas 300I Pro |            |

请在worker节点先安装NPU硬件对应的驱动和固件，才能构建昇腾NPU的训练和推理任务。安装文档[链接](https://support.huawei.com/enterprise/zh/category/ascend-computing-pid-1557196528909)。NPU驱动和固件、MindX DL平台组件、Toolbox的版本需要配套使用，请参阅官方文档获取配套的软件包

3. 存储（NFS、CephFS、OceanStore）

4. 容器镜像仓（harbor）

## 下载本工具

本工具只支持root用户，在master节点上运行。下载地址：[Ascend/ascend-hccl-controller](https://gitee.com/ascend/ascend-hccl-controller/tree/mindxdl-deploy/)。2种下载方式：

1. 使用git clone，切换到mindxdl-deploy分支

2. 下载mindxdl-deploy分支的[zip文件](https://gitee.com/ascend/ascend-hccl-controller/repository/archive/mindxdl-deploy.zip)

然后联系工程师取得离线依赖包resources.tar.gz（里面包括开源软件、镜像以及mindxdl的内置、预置镜像等），将离线依赖包解压在master节点的/root目录下。mindxdl的内置、预置镜像也需要跟NPU驱动和固件等配套使用。

```bash
root@master:~# ls
mindxdl-deploy
resources             //由resources.tar.gz解压得到，必须放置在/root目录下
resources.tar.gz
```

## 安装步骤

### 步骤1：安装ansible

工具中包含一个install_ansible.sh文件用于安装ansible

在工具目录中执行：

```bash
root@master:~/ascend-hccl-controller# bash install_ansible.sh

root@master:~/ascend-hccl-controller# ansible --version
config file = None
configured module search path = ['/root/.ansible/plugins/modules', '/usr/share/ansible/plugins/modules']
ansible python module location = /usr/local/lib/python3.6/dist-packages/ansible
ansible collection location = /root/.ansible/collections:/usr/share/ansible/collections
executable location = /usr/local/bin/ansible
python version = 3.6.9 (default, Jan 26 2021, 15:33:00) [GCC 8.4.0]
jinja version = 3.0.1
libyaml = True
```

ansible默认安装在python3（Ubuntu 18.04系统自带：python3.6.9，Centos 7.6：python3.6.8）中，安装完成后执行ansible --version查看ansible是否安装成功

注意：如果执行中报错“error: python3 must be python3.6 provided by the system by default, check it by run 'python3 -V'”，可能原因是环境上设置了相关环境变量或软连接，导致python3指向了其他的python版本，请保证python3命令指向系统自带的python3.6.9

### 步骤2：配置集群信息

在inventory_file文件中，需要提前规划好如下集群信息：

1. 指定安装harbor的服务器ip。默认为本机localhost，可更改为其他服务器ip

2. 当选用NFS存储方案时，指定安装nfs-server的服务器ip。默认为本机localhost，可更改为其他服务器ip。当【步骤3：配置安装信息】"STORAGE_TYPE"设置不为"NFS"时，此项配置无效，可删除

3. master节点ip。只能为本机localhost，不可更改

4. master_backup节点ip。默认无，即为单master集群。如需部署master高可用集群，这里至少需要配置2个或2个以上的节点ip（建议配置2个，即为3 master集群；本工具仅在3 master场景下经过全面测试）。不可包括master节点，即不可包括localhost。master_backup节点需要与master节点的系统架构一致

5. worker节点ip。默认无，即为无worker节点集群；可更改为其他服务器ip。如果这里包括master或master_backup组的ip，即把该ip的节点同时作为master和worker节点

```ini
[harbor]
localhost ansible_connection=local

[nfs_server]
localhost ansible_connection=local

[master]
localhost ansible_connection=local

[master_backup]

[worker]

# 这个默认配置，即把本机部署成一个单master节点的k8s集群，而且无worker节点
```

注意：

1. k8s要求集群内节点(master、worker、master_backup）的hostname不一样，因此建议执行安装前设置所有节点使用不同的hostname。如果未统一设置且存在相同hostname的节点，那么可在inventory_file文件中设置set_hostname主机变量，安装过程将自动设置节点的hostname。hostname需满足k8s和ansible的格式要求，建议用“[a-z]-[0-9]”的格式，如“worker-1”。例如：

2. 在部署master高可用集群时，必须给[master]和[master_backup]的节点设置kube_interface主机变量，以及增加一个[all:vars]的kube_vip主机组变量。kube_interface为各自节点实际使用的ip对应的网卡名称，可通过`ip a`查询，如"enp125s0f0"。kube_vip需跟k8s集群节点ip在同一子网，且为闲置、未被他人使用的ip，请联系网络管理员获取。

```ini
[harbor]
localhost ansible_connection=local

[nfs_server]
localhost ansible_connection=local

[master]
localhost ansible_connection=local  set_hostname="master"  kube_interface="enp125s0f0"

[master_backup]
192.0.3.100  set_hostname="master-backup-1"  kube_interface="enp125s0f0"
192.0.3.101  set_hostname="master-backup-2"  kube_interface="enp125s0f0"

[worker]
192.0.2.50  set_hostname="worker-1"
192.0.2.51  set_hostname="worker-2"
192.0.2.52  set_hostname="worker-3"

[all:vars]
kube_vip="192.0.4.200"

# 这个配置，即部署一个3 master高可用k8s集群，而且有3个worker节点
# 以上192.0.*.*等ip仅为示例，请修改为实际规划的ip地址
```

3. （高阶）如果需要配置管理面网络和业务面网络分离，各节点硬件上需要2张网卡。必须给[master]和[master_backup]设置kube_interface和apiserver_advertise_address主机变量，分别为各自节点业务面网卡名称和业务面网络ip；[worker]需设置kube_interface。[all:vars]的kube_vip主机组变量为业务面网络的闲置ip，harbor_host_ip主机组变量为harbor主机的业务面网络ip

```ini
[harbor]
localhost ansible_connection=local

[nfs_server]
localhost ansible_connection=local

[master]
localhost ansible_connection=local  set_hostname="master"  kube_interface="enp125s0f1"  apiserver_advertise_address="195.0.3.99"

[master_backup]
192.0.3.100  set_hostname="master-backup-1"  kube_interface="enp125s0f1"  apiserver_advertise_address="195.0.3.100"
192.0.3.101  set_hostname="master-backup-2"  kube_interface="enp125s0f1"  apiserver_advertise_address="195.0.3.101"

[worker]
192.0.2.50  set_hostname="worker-1"  kube_interface="enp125s0f1"
192.0.2.51  set_hostname="worker-2"  kube_interface="enp125s0f1"
192.0.2.52  set_hostname="worker-3"  kube_interface="enp125s0f1"

[all:vars]
kube_vip="195.0.4.200"
harbor_host_ip="195.0.3.99"

# 这个配置，即部署一个管理面网络和业务面网络分离的3 master高可用k8s集群，而且有3个worker节点
# 以上192.0.*.*、195.0.*.*等ip仅为示例，请修改为实际规划的ip地址
# 192.0.*.*为管理面网络，195.0.*.*为业务面网络；kube_interface、apiserver_advertise_address分别为业务面网络的网卡名称和ip；kube_vip在业务面网络中，且为闲置、未被他人使用的ip；harbor_host_ip为harbor主机的业务面网络ip
# 在多网卡的复杂网络场景，建议把master和master_backup的apiserver_advertise_address主机变量、all的harbor_host_ip主机组变量都配置上
```


inventory_file文件配置详细可参考[[How to build your inventory &mdash; Ansible Documentation](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html)]

### 步骤3：配置安装信息

在group_vars目录中的all.yaml文件

```yaml
# harbor https port
HARBOR_HTTPS_PORT: 7443
# harbor install path
HARBOR_PATH: /data/harbor
# password for harbor, can not be empty, delete immediately after finished
HARBOR_PASSWORD: ""

# password for mysql, can not be empty, delete immediately after finished
MYSQL_PASSWORD: ""

# password for redis, can not be empty, delete immediately after finished
REDIS_PASSWORD: ""

# select "NFS" or "CEPHFS" or "OCEANSTORE" as the storage solution, default to "NFS"
STORAGE_TYPE: "NFS"
# mindx-dl platform storage path on "NFS" or "CEPHFS" or "OCEANSTORE", default to /data/atlas_dls
STORAGE_PATH: "/data/atlas_dls"
# storage capatity, default to "5120Gi", i.e. 5Ti
STORAGE_CAPACITY: "5120Gi"

# cephfs monitor ip. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_IP: ""
# cephfs port. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_PORT: ""
# cephfs user. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_USER: ""
# cephfs key. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_KEY: ""

# k8s pod network cidr
POD_NETWORK_CIDR: "192.168.0.0/16"
# k8s service cidr
SERVICE_CIDR: "10.96.0.0/12"
# mindx k8s namespace
K8S_NAMESPACE: "mindx-dl"

#mindx user
MINDX_USER: hwMindX
MINDX_USER_ID: 9000
MINDX_GROUP: hwMindX
MINDX_GROUP_ID: 9000

#HwHiAiUser group
HIAI_GROUP: HwHiAiUser
```

其中中配置项详细为：

| 配置项               | 说明                                   |
| ----------------- | ------------------------------------ |
| HARBOR_HTTPS_PORT | harbor的https监听端口，默认为7443             |
| HARBOR_PATH       | Harbor的安装路径，默认为/data/harbor                   |
| HARBOR_PASSWORD   | harbor的登录密码，不可为空，**必须配置**。**安装完成后应立即删除** |
| MYSQL_PASSWORD    | mysql的登录密码，不可为空，**必须配置**。**安装完成后应立即删除**  |
| REDIS_PASSWORD    | redis的登录密码，不可为空，**必须配置**。**安装完成后应立即删除**  |
| STORAGE_TYPE      | 由用户按需选用的存储方案，默认为"NFS"；也可选"CEPHFS"或"OCEANSTORE"           |
| STORAGE_PATH      | 存储的共享路径，默认为/data/atlas_dls   |
| STORAGE_CAPACITY  | 存储的共享容量，默认为5Ti，**请根据实际配置**   |
| CEPHFS_IP         | cephfs集群的monitor ip，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*  |
| CEPHFS_PORT       | cephfs集群的port，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*  |
| CEPHFS_USER       | cephfs集群的用户名，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*。一般为admin  |
| CEPHFS_KEY        | cephfs集群的密钥，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*。可在cephfs monitor节点通过`ceph auth get-key client.admin`查询。**安装完成后应立即删除**  |
| POD_NETWORK_CIDR  | k8s默认pod网段，不可跟其他ip网段重叠或冲突            |
| SERVICE_CIDR      | k8s默认service网段，不可跟其他ip网段重叠或冲突        |
| K8S_NAMESPACE     | mindx dl组件默认k8s命名空间                  |
| MINDX_USER        | mindx dl组件默认运行用户                     |
| MINDX_USER_ID     | mindx dl组件默认运行用户id                   |
| MINDX_GROUP       | mindx dl组件默认运行用户组                    |
| MINDX_GROUP_ID    | mindx dl组件默认运行用户组id                  |
| HIAI_GROUP        | 驱动默认运行用户组                    |

注：

1. harbor的登录用户名默认为admin。

2. 本工具支持使用NFS、CephFS、OceanStore 3种存储方案，默认选用NFS方案。

   - 3.1 当"STORAGE_TYPE"配置项为"NFS"时，请确认【步骤2：配置集群信息】inventory_file的"nfs_server"配置正确。

   - 3.2 当"STORAGE_TYPE"配置项为"CEPHFS"时，请提前准备好cephfs集群，并确认"CEPHFS_IP"、"CEPHFS_PORT"、"CEPHFS_USER"、"CEPHFS_KEY"这4个配置项填写正确。

   - 3.3 当"STORAGE_TYPE"配置项为"OCEANSTORE"时，请提前准备好oceanstore集群。

3. 使用CephFS或OceanStore方案时，需要手动挂载并在挂载目录下创建STORAGE_PATH（默认为/data/atlas_dls）目录及其下的相关目录，并修改该目录属主为hwMindX用户。具体操作请参考tools/create_storage_dir.sh。

4. k8s默认使用"192.168.0.0/16"和"10.96.0.0/12"分别作为内部的pod和service网段，不可跟其他网段重叠或冲突。请规划好集群内的ip资源，必要时可根据实际修改POD_NETWORK_CIDR和SERVICE_CIDR配置项

### 步骤4：检查集群状态

如果inventory_file内配置了非localhost的远程ip，根据ansible官方建议，请用户自行使用SSH密钥的方式连接到远程机器，可参考[[connection_details; Ansible Documentation](https://docs.ansible.com/ansible/latest/user_guide/connection_details.html#setting-up-ssh-keys)]

在工具目录中执行：

```bash
root@master:~/ascend-hccl-controller# ansible -i inventory_file all -m ping

localhost | SUCCESS => {
    "ansible_facts": {
        "discovered_interpreter_python": "/usr/bin/python3"
    },
    "changed": false,
    "ping": "pong"
}
worker1_ipaddres | SUCCESS => {
    "ansible_facts": {
        "discovered_interpreter_python": "/usr/bin/python3"
    },
    "changed": false,
    "ping": "pong"
}
```

当所有节点都能ping通，则表示inventory_file文件中所有节点连通性正常。否则，请检查节点的ssh连接和inventory_file文件配置是否正确

各个节点应保持时间同步，不然可能会出现不可预知异常，时间同步服务应当由网络管理员提供支持。无法获取网络管理员提供支持的时间同步服务时，本工具也提供了可选的时间同步服务，依次执行01、99这2个子任务即可。

```bash
root@master:~/ascend-hccl-controller# ansible-playbook -i inventory_file playbooks/01.resource.yaml playbooks/99.chrony.yaml
```

### <a name="resources_no_copy">步骤5：执行安装</a>

在工具目录中执行：

```bash
root@master:~/ascend-hccl-controller# ansible-playbook -i inventory_file all.yaml
```

注：

1. k8s节点不可重复初始化或加入，执行本步骤前，请先在master和worker节点执行如下命令，清除节点上已有的k8s系统
   ```bash
   kubeadm reset -f; iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X; systemctl restart docker
   ```

   3 master场景下，执行以上命令后，3个master节点上可能会残留inventory_file中设置的“kube-vip”，绑定在“kube_interface”上。需要在3个master节点上确认并手动删除
   ```bash
   ip a  # 查询网卡（kube_interface）上是否存在此ip（kube-vip）
   ip addr delete <kube-vip> dev <kube_interface>  # 如果存在，需手动删除
   ```

2. mysql数据库会持久化MindX DL平台组件的相关数据，存放在外部存储目录（/data/atlas_dls/platform/mysql）。如果需手动清除k8s系统，请务必删除这个目录，避免后续MindX DL平台组件运行异常。详细目录见下。

外部存储中的MindX DL平台目录结构
   ```bash
   |── atlas_dls  # 平台总目录
      |── platform  # 平台组件相关目录
         ├── kmc  # DL组件的kmc证书目录
            |── alarm-manager
            |── apigw
            ...
         ├── log  # DL组件的log目录
            |── alarm-manager
            |── apigw
            ...
         ├── mysql  # mysql的安装目录
         ├── services  # DL组件的私有配置目录
      |── group1  # 平台组织id1
         |── user1  # 平台用户id1
         |── user2  # 平台用户id2
         ...
      |── group2  # 平台组织id2
         |── user3  # 平台用户id3
         |── user4  # 平台用户id4
         |── user5  # 平台用户id5
         ...
      ...
   ```

3. 如果某节点docker.service配置了代理，则可能无法访问harbor镜像仓。使用本工具前，请先在`/etc/systemd/system/docker.service.d/proxy.conf`中NO_PROXY添加harbor host的ip，然后执行`systemctl daemon-reload && systemctl restart docker`生效

4. 如果inventory_file内配置了非localhost的远程ip，本工具会将本机的/root/resources目录分发到远程机器上。如果有重复执行以上命令的需求，可在以上命令后加`-e resources_no_copy=true`参数，避免重复执行耗时的~/resources目录打包、分发操作

### 步骤6：安装后检查

检查kubernetes节点

```bash
root@master:~# kubectl get nodes -A
NAME             STATUS   ROLES    AGE   VERSION
master           Ready    master   60s   v1.19.16
worker-1         Ready    worker   60s   v1.19.16
```

检查kubenetes pods

```bash
root@master:~# kubectl get pods -A
NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
kube-system   calico-kube-controllers-659bd7879c-l7q55   1/1     Running   2          19h
kube-system   calico-node-5zk76                          1/1     Running   1          19h
kube-system   calico-node-cxhdn                          1/1     Running   0          19h
kube-system   coredns-f9fd979d6-l42rb                    1/1     Running   2          19h
kube-system   coredns-f9fd979d6-x2bg2                    1/1     Running   2          19h
kube-system   etcd-node-10-0-2-15                        1/1     Running   1          19h
kube-system   kube-apiserver-node-10-0-2-15              1/1     Running   1          19h
kube-system   kube-controller-manager-node-10-0-2-15     1/1     Running   5          19h
kube-system   kube-proxy-g65rn                           1/1     Running   1          19h
kube-system   kube-proxy-vqzb7                           1/1     Running   0          19h
kube-system   kube-scheduler-node-10-0-2-15              1/1     Running   4          19h
mindx-dl      alertmanager-5675466341-45789              1/1     Running   1          19h
mindx-dl      grafana-core-58664d599b-4d8s8              1/1     Running   1          19h
mindx-dl      kube-state-metrics-592645991b-2f7s5        1/1     Running   1          19h
mindx-dl      mysql-55569fc484-bb6kw                     1/1     Running   1          19h
mindx-dl      node-exporter-ds5f5                        1/1     Running   0          19h
mindx-dl      node-exporter-s5j9s                        1/1     Running   1          19h
mindx-dl      prometheus-577fb6b799-k6mwl                1/1     Running   1          19h
mindx-dl      redis-deploy-85dbb68c56-cfxhq              1/1     Running   1          19h
```

注：

1. 手动执行kubectl命令时，需取消http(s)_proxy系统网络代理配置，否则会连接报错或一直卡死

### 步骤7：安装MindX DL平台组件

1. 在~/resources/目录下创建mindxdl目录。如果该目录已存在，请确保该目录下为空

   ```bash
      mkdir -p ~/resources/mindxdl
   ```

2. 将MindX DL平台组件的版本发布件放到~/resource/mindxdl目录中。只需要放master节点CPU架构的包即可

   ```bash
   ~/resources/
    `── mindxdl
        ├── Ascend-mindxdl-apigw_{version}-{arch}.zip
        ├── Ascend-mindxdl-cluster-manager_{version}-{arch}.zip
         ....
   ```

3. 在工具目录中执行安装命令

   ```bash
   root@master:~/ascend-hccl-controller# ansible-playbook -i inventory_file playbooks/16.mindxdl.yaml
   ```

4. （可选）如果k8s集群中包含跟master节点的CPU架构不一致的worker节点，则需要单独执行这一步，用来构建npu-exporter、device-plugin镜像。

   4.1 任意选择在某个异构的worker节点，将对应CPU架构的npu-exporter、device-plugin组件放到某个目录，比如/tmp/mindxdl；将本工具的tools/build_image.sh构建脚本也传到这个目录

   ```bash
   /tmp/
    `── mindxdl
      ├── build_image.sh
      ├── Ascend-mindxdl-npu-exporter_{version}-{arch}.zip
      ├── Ascend-mindxdl-device-plugin_{version}-{arch}.zip
   ```

   4.2 在该worker节点，执行如下命令，用来构建镜像并上传到harbor。\<harbor-ip\>和\<harbor-port\>分别为之前部署的harbor仓的ip和端口（端口默认为7443），\<*zip_file\>为上一步传上去的npu-exporter或device-plugin的zip包路径
   ```bash
   root@worker-1:/tmp/mindxdl# bash build_image.sh --harbor-ip <harbor-ip> --harbor-port <harbor-port> <npu-exporter_zip_file>

   root@worker-1:/tmp/mindxdl# bash build_image.sh --harbor-ip <harbor-ip> --harbor-port <harbor-port> <device-plugin_zip_file>
   ```

   4.3 回到master节点的本工具目录，执行如下命令，将构建的harbor镜像拉取到所有的worker节点。\<tag\>为镜像tag，可通过组件包内的yaml查询。跟上一步构建镜像的worker节点的架构不一致的其他worker节点，镜像会拉取失败，但无影响，因为镜像已存在
   ```bash
   root@master:~/ascend-hccl-controller# ansible worker -i inventory_file -m shell -a "docker pull <harbor-ip>:<harbor-port>/mindx/ascend-k8sdeviceplugin:<tag>"

   root@master:~/ascend-hccl-controller# ansible worker -i inventory_file -m shell -a "docker pull <harbor-ip>:<harbor-port>/mindx/npu-exporter:<tag>"
   ```

注：

1. MindX DL平台组件安装时依赖harbor。安装过程会制作镜像并上传到harbor中

2. 只支持安装MindX DL平台组件，当前包括14个平台组件（apigw、cluster-manager、data-manager、dataset-manager、image-manager、model-manager、inference-manager、train-manager、user-manager、alarm-manager、hccl-controller、volcano、npu-exporter、device-plugin）。其中npu-exporter、device-plugin部署在worker节点，其他组件都部署在master节点

3. npu-exporter、device-plugin组件包内的部分版本由于安全整改，可能没有Dockerfile和yaml文件，需要获取到对应版本的文件并重新打包，获取地址：[链接](https://gitee.com/ascend/mindxdl-deploy/tags)。NPU驱动和固件、MindX DL平台组件、Toolbox的版本需要配套使用，请参阅官方文档获取配套的软件包

### 步骤8：安装Ascend-Docker-Runtime组件

Ascend-Docker-Runtime组件包含在MindX Toolbox包中，需要先获取MindX Toolbox包

1. 在~/resources/目录下创建mindx-toolbox目录。如果该目录已存在，请确保该目录下为空

   ```bash
      mkdir -p ~/resources/mindx-toolbox
   ```

2. 将MindX Toolbox包放到~/resource/mindx-toolbox目录中。需要放worker节点CPU架构的包

   ```bash
   ~/resources/
    `── mindx-toolbox
        ├── Ascend-mindxdl-toolbox_{version}-{arch}.zip
         ....
   ```

3. 在工具目录中执行安装命令。MindX Toolbox中的Ascend-Docker-Runtime即可安装到各个worker节点

   ```bash
   root@master:~/ascend-hccl-controller# ansible-playbook -i inventory_file playbooks/17.mindx-toolbox.yaml
   ```

## 更新MindX DL平台组件

如果用户已完整执行过以上安装步骤，本工具支持单独更新MindX DL平台组件

1. 查阅“步骤2：配置集群信息”的inventory_file文件和“步骤3：配置安装信息”的group_vars/all.yaml文件，确保这2个配置文件同上一次使用本工具时的配置完全一致

2. 执行“步骤7：安装MindX DL平台组件”。该步骤可重复执行

## 安装后操作

如果worker节点中包含中心训练硬件时，需要配置device的网卡IP。具体操作参考[[配置device的网卡IP](https://support.huawei.com/enterprise/zh/doc/EDOC1100234042/5a225af5)]

## 分步骤安装

playbooks目录下有很多文件，其中每个yaml文件对应一个组件，可以实现只安装某个组件

```bash
playbooks/
├── 01.resource.yaml  # 分发/root/resources目录（耗时较长）
├── 02.docker.yaml  # 安装docker
├── 03.harbor.yaml  # 安装harbor并登录
├── 04.open-source-image.yaml  # 推送/root/resources/images里的开源镜像到harbor（耗时较长）
├── 05.basic.yaml  # 创建MindX DL所需的用户、日志目录等基础操作
├── 06.k8s.yaml  # 安装k8s系统
├── 07.nfs.yaml  # 安装nfs并创建nfs的pv。当"STORAGE_TYPE"设置不为"NFS"时，此步骤会自动跳过
├── 08.cephfs.yaml  # 创建cephfs的pv、secret。当"STORAGE_TYPE"设置不为"CEPHFS"时，此步骤会自动跳过
├── 09.oceanstore.yaml  # 创建oceanstore的pv（hostpath方式)。当"STORAGE_TYPE"设置不为"OCEANSTORE"时，此步骤会自动跳过
├── 10.pvc.yaml  # 创建pvc
├── 11.mysql.yaml  # 安装mysql
├── 12.redis.yaml  # 安装redis
├── 13.prometheus.yaml  # 安装prometheus、grafana、node-exporter、alertmanager、kube-state-metrics
├── 14.inner-image.yaml  # 推送/root/resources/mindx-inner-images里的内置镜像到harbor（耗时较长）
├── 15.pre-image.yaml  # 推送/root/resources/mindx-pre-images里的预置镜像到harbor（耗时较长）
├── 16.mindxdl.yaml  # 安装或更新MindX DL平台组件和基础组件
├── 17.mindx-toolbox.yaml  # 安装或更新MindX Toolbox
├── 99.chrony.yaml  # （可选）安装chrony
```

例如:

1. 只分发软件包，则执行

   ```bash
   ansible-playbook -i inventory_file playbooks/01.resource.yaml
   ```

   可在以上命令后加`-e resources_no_copy=true`参数，该参数作用请见<a href="#resources_no_copy">步骤5：执行安装注意事项第4点</a>

2. 只安装k8s系统，则执行

   ```bash
   ansible-playbook -i inventory_file playbooks/06.k8s.yaml
   ```

   k8s节点不可重复初始化或加入，执行本步骤前，请先在master和worker节点执行如下命令，清除节点上已有的k8s系统
   ```bash
   kubeadm reset -f; iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X; systemctl restart docker
   ```

   3 master场景下，执行以上命令后，3个master节点上可能会残留inventory_file中设置的“kube-vip”，绑定在“kube_interface”上。需要在3个master节点上确认并手动删除
   ```bash
   ip a  # 查询网卡（kube_interface）上是否存在此ip（kube-vip）
   ip addr delete <kube-vip> dev <kube_interface>  # 如果存在，需手动删除
   ```

   由于ansible的幂等性，除playbooks/06.k8s.yaml步骤外，其他步骤均可以重复执行

3. 工具目录下的all.yaml为全量安装，安装效果跟依次执行playbooks目录下的01~15编号的yaml效果一致（不包括16.mindxdl.yaml、17.mindx-toolbox.yaml、99.chrony.yaml）。实际安装时可根据需要对组件灵活删减

   如果需要重新部署DL平台，手动清除k8s系统及DL平台残留的mysql数据库目录后，只需分别依次执行06-16这些子任务（这些子任务都跟k8s相关）即可，不必执行01-05、17这些子任务

4. （可选）各个节点应保持时间同步，不然可能会出现不可预知异常，时间同步服务应当由网络管理员提供支持。无法获取网络管理员提供支持的时间同步服务时，本工具也提供了可选的时间同步服务，执行01、99这2个子任务即可。

```bash
root@master:~/ascend-hccl-controller# ansible-playbook -i inventory_file playbooks/01.resource.yaml playbooks/99.chrony.yaml
```

# FAQ

1. Q: 某个节点的calico-node-**出现READY “0/1”，`kubectl describe pod calico-node-**(master的calico-node)`时有报错信息“calico/node is not ready: BIRD is not ready: BGP not established with \<ip\>”

- A: 可能是该节点的交换分区被打开了（swap on，可通过`free`查询)，kubelet日志报错“failed to run Kubelet: running with swap on is not supported, please disable swap”，导致该节点calico访问失败。解决方案是禁用swap（执行`swapoff -a`）

2. Q: 安装某组件时报错，报错信息中包含访问harbor镜像仓失败等字样

- A: harbor镜像仓会管理安装过程中的所有镜像。首先可能是某节点docker.service配置了代理，具体请见<a href="#resources_no_copy">步骤5：执行安装注意事项第3点</a>。其次可能是harbor服务异常，可在harbor主机上执行`docker ps | grep goharbor`，如果不是存在9个容器且均为up状态，可执行`systemctl restart harbor`重启harbor服务。如果重启harbor服务后harbor服务仍然异常，建议直接重装harbor（执行04.harbor.yaml子任务）

3. 初始化master时，有报错信息“no default routes found in /proc/net/route or /proc/net/ipv6_route”、“cannot use 0.0.0.0 as the bind address for the API Server"等

- A: 多网卡的复杂网络场景下，可能会出现这个问题。请确认网卡路由是否畅通

4. 3 master场景下，2个master_backup节点执行加入k8s集群命令时，报错访问\<kube-vip\>:6443被拒。

- A: 可能是该节点上残留之前部署的“kube-vip”，手动删除该ip即可。具体操作可见上文

5. k8s集群部署起来后，pod ip和node ip不是预期的业务ip，而是其他无关的ip

- A: 多网卡的复杂网络场景下，可能会出现这个问题。建议在各个节点上配置kubelet参数，操作如下
   ```bash
   vi /etc/default/kubelet  # 新建这个配置文件
   KUBELET_EXTRA_ARGS="--node-ip=195.0.3.99"  # 上面的新建的文件中写入此内容，195.0.3.99请改为该节点预期的业务ip

   systemctl daemon-reload && systemctl restart kubelet  # 重启kubelet，使配置生效
   ```

6. coredns在centos7.6上可能报错"plugin/forward: no nameservers found"

- A: 修改configmap coredns，操作如下

   ```bash
   kubectl edit configmap coredns -n kube-system  # 进入configmap coredns的编辑状态

   将里面的"forward . /etc/resolv.conf" 改成 "forward . 8.8.8.8"
   ```

7. 3 master场景下，任意一个master节点宕机后，一般会等待约6分钟，宕机节点的pod才会迁移到其他可用节点。在这段约6分钟的pod迁移时间内，mindxdl平台将不可用。
