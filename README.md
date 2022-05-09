# 快速入门

## 功能描述

本文主要介绍如何使用ansible安装mindxdl所需开源软件安装。其中包含如下开源软件

| 软件名        | 备注                    |
| ---------- | -------------------------|
| docker     | 集群中所有节点都需要安装   |
| harbor     | 容器镜像仓                |
| kubernetes | k8s平台                  |
| nfs        | nfs存储系统              |
| mysql      | 安装在k8s集群中，关系型数据库系统 |
| redis      | 安装在k8s集群中，非关系型数据库系统 |
| kubeedge   | 安装在k8s集群中，使能边缘计算的平台 |
| prometheus + grafana + node-exporter | 安装在k8s集群中，资源监控组件      |

## 环境要求

### 支持的操作系统说明

| 操作系统   | 版本   | CPU架构 |
|:------:|:---------:|:-------:|
| Ubuntu | 18.04.1/5 | aarch64 |
| Ubuntu | 18.04.1/5 | x86_64  |

根目录的磁盘空间利用率高于85%会触发Kubelet的镜像垃圾回收机制，将导致服务不可用。请确保根目录有足够的磁盘空间，建议大于1 TB

### 支持的硬件形态说明

| 中心推理硬件    | 中心训练硬件|
|:--------------:|:----------:|
| A300-3000      | A300T-9000 |
| A300-3010      | A800-9000  |
| Atlas 300I Pro | A800-9010  |
| A800-3000      |            |
| A800-3010      |            |

## 下载本工具

本工具只支持root用户，下载地址：[MindXDL-deploy: MindX DL platform deployment](https://gitee.com/ascend/mindxdl-deploy)。2种下载方式：

1. 使用git clone

2. 下载master分支的[zip文件](https://gitee.com/ascend/mindxdl-deploy/repository/archive/master.zip)

然后联系工程师取得开源软件的resources.tar.gz离线安装包，将离线安装包解压在/root目录下。按如下方式放置

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
root@master:~/mindxdl-deployer# bash install_ansible.sh

root@master:~/mindxdl-deployer# ansible --version
config file = None
configured module search path = ['/root/.ansible/plugins/modules', '/usr/share/ansible/plugins/modules']
ansible python module location = /usr/local/lib/python3.6/dist-packages/ansible
ansible collection location = /root/.ansible/collections:/usr/share/ansible/collections
executable location = /usr/local/bin/ansible
python version = 3.6.9 (default, Jan 26 2021, 15:33:00) [GCC 8.4.0]
jinja version = 3.0.1
libyaml = True
```

ansible默认安装在系统自带python3（Ubuntu：python3.6.9）中，安装完成后执行ansible --version查看ansible是否安装成功

注意：如果执行中报错“error: python3 must be python3.6 provided by the system by default, check it by run 'python3 -V'”，可能原因是环境上设置了相关环境变量或软连接，导致python3指向了其他的python版本，请保证python3命令指向系统自带的python3.6.9

### 步骤2：配置集群信息

在inventory文件中，需要提前规划好如下集群信息：

1. 安装harbor的服务器ip

2. master节点ip，只能为本机localhost

3. work节点ip（默认无，请根据需要添加，不可包括master节点，即不可包括localhost）

4. mysql安装的节点ip，只能为本机localhost

5. nfs服务器ip。nfs可使用已有nfs服务器。当【步骤3：配置安装信息】"STORAGE_TYPE"设置为"CEPHFS"时，此项配置无效，可删除。

```bash
[harbor]
localhost ansible_connection=local

[master]
localhost ansible_connection=local

[worker]
worker1_ip
worker2_ip
worker3_ip

[mysql]
localhost ansible_connection=local

[nfs_server]
localhost ansible_connection=local
```

注意：k8s要求所有设备的hostname不一样，因此建议执行安装前设置所有设备使用不同的hostname。如果未统一设置且存在相同hostname的设备，那么可在inventory文件中设置set_hostname变量，安装过程将自动设置设备的hostname。hostname需满足k8s和ansible的格式要求，建议用“[a-z]-[0-9]”的格式，如“worker-1”。例如：

```ini
[master]
localhost ansible_connection=local

[worker]
worker1_ipaddress  set_hostname="worker-1"
worker2_ipaddress  set_hostname="worker-2"
worker3_ipaddress
```

inventory文件配置详细可参考[[How to build your inventory &mdash; Ansible Documentation](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html)]

### 步骤3：配置安装信息

在group_vars目录中的all.yaml文件

```yaml
# harbor ip address
HARBOR_HOST_IP: ""
# harbor https port
HARBOR_HTTPS_PORT: 7443
# harbor install path
HARBOR_PATH: /data/harbor
# password for harbor, can not be empty, delete immediately after finished
HARBOR_PASSWORD: ""

# mysql install path
MYSQL_DATAPATH: /data/mysql
# password for mysql, can not be empty, delete immediately after finished
MYSQL_PASSWORD: ""

# select "NFS" or "CEPHFS" as the storage solution, default to "NFS"
STORAGE_TYPE: "NFS"

# nfs shared path, can be multiple configurations. can not be empty if "STORAGE_TYPE" is "NFS"
NFS_PATH: ["/data/atlas_dls"]

# cephfs monitor ip. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_IP: ""
# cephfs port. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_PORT: ""
# cephfs user. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_USER: ""
# cephfs key. can not be empty if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_KEY: ""
# cephfs request storage. can not be zero if "STORAGE_TYPE" is "CEPHFS"
CEPHFS_REQUEST_STORAGE: "0Gi"

# mindx k8s namespace
K8S_NAMESPACE: "mindx-dl"
# ip address for api-server
K8S_API_SERVER_IP: ""

#mindx user
MINDX_USER: hwMindX
MINDX_USER_ID: 9000
MINDX_GROUP: hwMindX
MINDX_GROUP_ID: 9000
```

其中中配置项详细为：

| 配置项               | 说明                                   |
| ----------------- | ------------------------------------ |
| HARBOR_HOST_IP    | 配置harbor的监听ip，多网卡场景下*建议配置*         |
| HARBOR_HTTPS_PORT | harbor的https监听端口，默认为7443             |
| HARBOR_PATH       | Harbor的安装路径，默认为/data/harbor                   |
| HARBOR_PASSWORD   | harbor的登录密码，不可为空，**必须配置**。**安装完成后应立即删除** |
| MYSQL_DATAPATH    | mysql的安装路径，默认为/data/mysql                           |
| MYSQL_PASSWORD    | mysql的登录密码，不可为空，**必须配置**。**安装完成后应立即删除**  |
| STORAGE_TYPE      | 由用户按需选用的存储方案，默认为"NFS"；也可选"CEPHFS"           |
| NFS_PATH          | nfs服务器的共享路径，可配置多个路径，默认为/data/atlas_dls。   |
| CEPHFS_IP         | cephfs集群的monitor ip，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*  |
| CEPHFS_PORT       | cephfs集群的port，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*  |
| CEPHFS_USER       | cephfs集群的用户名，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*  |
| CEPHFS_KEY        | cephfs集群的密钥，*"STORAGE_TYPE"设置为"CEPHFS"时不可为空，必须配置*。**安装完成后应立即删除**  |
| CEPHFS_REQUEST_STORAGE| cephfs集群分配的存储空间，*"STORAGE_TYPE"设置为"CEPHFS"时不可为"0Gi"，必须配置*。  |
| K8S_NAMESPACE     | mindx dl组件默认k8s命名空间                  |
| K8S_API_SERVER_IP | K8s的api server监听地址，多网卡场景下*建议配置*    |
| MINDX_USER        | mindx dl组件默认运行用户                     |
| MINDX_USER_ID     | mindx dl组件默认运行用户id                   |
| MINDX_GROUP       | mindx dl组件默认运行用户组                    |
| MINDX_GROUP_ID    | mindx dl组件默认运行用户组id                  |

注：

1. harbor的登录用户名默认为admin

2. 本工具支持使用nfs和cephfs 2种存储方案，默认选用nfs方案。用户可通过设置"STORAGE_TYPE"为"CEPHFS"选用cephfs方案。

   - 2.1 当"STORAGE_TYPE"配置项为"NFS"时，请确认"NFS_PATH"配置项和【步骤2：配置集群信息】inventory的"nfs_server"配置正确。

   - 2.2 当"STORAGE_TYPE"配置项为"CEPHFS"时，请提前准备好cephfs集群，并确认"CEPHFS_IP"、"CEPHFS_PORT"、"CEPHFS_USER"、"CEPHFS_KEY"、"CEPHFS_REQUEST_STORAGE"这5个配置项填写正确。

### 步骤4：检查集群状态

如果inventory_file内配置了非localhost的远程ip，根据ansible官方建议，请用户自行使用SSH密钥的方式连接到远程机器，可参考[[connection_details; Ansible Documentation](https://docs.ansible.com/ansible/latest/user_guide/connection_details.html#setting-up-ssh-keys)]

在工具目录中执行：

```bash
root@master:~/mindxdl-deployer# ansible -i inventory_file all -m ping

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

当所有设备都能ping通，则表示inventory中所有设备连通性正常。否则，请检查设备的ssh连接和inventory文件配置是否正确

### <a name="resources_no_copy">步骤5：执行安装</a>

在工具目录中执行：

```bash
root@master:~/mindxdl-deployer# ansible-playbook -i inventory_file all.yaml
```

注：

1. k8s节点不可重复初始化或加入，使用本工具前，请先在master和worker节点执行`kubeadm reset`清除节点上已有的k8s系统

2. mysql数据库会持久化MindX DL平台组件的相关数据，存储在master节点的MYSQL_DATAPATH目录下（默认为/data/mysql）。如果需手动清除k8s系统，请务必也删除该目录，避免后续MindX DL平台组件运行异常

3. 如果docker.service配置了代理，则可能无法访问harbor镜像仓。使用本工具前，请先在`/etc/systemd/system/docker.service.d/proxy.conf`中NO_PROXY添加harbor host的ip，然后执行`systemctl daemon-reload && systemctl restart docker`生效

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
mindx-dl      grafana-core-58664d599b-4d8s8              1/1     Running   1          19h
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

2. 将MindX DL平台组件放到~/resource/mindxdl目录中
   
   ```bash
   ~/resources/
    `── mindxdl
        ├── Ascend-mindxdl-apigw_{version}-{arch}.zip
        ├── Ascend-mindxdl-cluster-manager_{version}-{arch}.zip
         ....
   ```

3. 在工具目录中执行安装命令

   ```bash
   root@master:~/mindxdl-deployer# ansible-playbook -i inventory_file playbooks/13.mindxdl.yaml
   ```

注：

1. MindX DL平台组件安装时依赖harbor。安装过程会制作镜像并上传到harbor中

2. 只支持安装MindX DL平台组件，当前包括11个平台组件（apigw、cluster-manager、data-manager、dataset-manager、edge-manager、image-manager、label-manager、model-manager、task-manager、train-manager、user-manager）

## 更新MindX DL平台组件

如果用户已完整执行过以上安装步骤，本工具支持单独更新MindX DL平台组件

1. 查阅“步骤2：配置集群信息”的inventory文件和“步骤3：配置安装信息”的group_vars/all.yaml文件，确保这2个配置文件同上一次使用本工具时的配置完全一致

2. 执行“步骤7：安装MindX DL平台组件”。该步骤可重复执行

## 后续依赖软件安装

完成安装MindX DL平台后，还需要完成相关依赖软件的安装，才能构建昇腾NPU的训练和推理任务

| 软件名称         | 安装文档     |
|:---------------:|:-----------:|
| NPU的驱动和固件  | 请参见各硬件产品中驱动和固件安装升级指南获取对应的指导 | 
| Ascend docker runtime | 请参见MindX 《MindX ToolBox用户指南》安装实用工具包“Ascend-mindx-toolbox_{version}_linux-{arch}.run” |
| MindX DL基础组件 | 请参见MindX 《MindX DL用户指南》安装4个基础组件（Volcano、Ascend Device Plugin、NPU-Exporter、HCCL-Controller） |

上述软件安装文档[链接](https://support.huawei.com/enterprise/zh/category/ascend-computing-pid-1557196528909)

# 详细说明

## 分步骤安装

playbooks目录下有很多文件，其中每个yaml文件对应一个组件，可以实现只安装某个组件

```bash
playbooks/
├── 01.resource.yaml  # 分发/root/resources目录
├── 02.basic.yaml  # 创建MindX DL所需的用户、日志目录等基础操作
├── 03.docker.yaml  # 安装docker
├── 04.harbor.yaml  # 安装harbor并登录
├── 05.open-source-image.yaml  # 推送/root/resources/images里的开源镜像到harbor
├── 06.k8s.yaml  # 安装k8s系统
├── 07.mysql.yaml  # 安装mysql
├── 08.redis.yaml  # 安装redis
├── 09.prometheus.yaml  # 安装prometheus、grafana、node-exporter
├── 10.kubeedge.yaml  # 安装kubeedge
├── 11.nfs.yaml  # 安装nfs
├── 12.cephfs.yaml  # 创建cephfs的pv、pvc、secret
├── 13.inner-image.yaml  # 推送/root/resources/mindx-inner-images里的内置镜像到harbor
├── 14.pre-image.yaml  # 推送/root/resources/mindx-pre-images里的预置镜像到harbor
├── 15.mindxdl.yaml  # 安装或更新MindX DL平台组件
```

例如:

1. 分发软件包
   
   ```bash
   ansible-playbook -i inventory_file playbooks/01.resource.yaml
   ```

   可在以上命令后加`-e resources_no_copy=true`参数，该参数作用请见<a href="#resources_no_copy">步骤5：执行安装注意事项第4点</a>

2. 只安装k8s系统，则执行
   
   ```bash
   ansible-playbook -i inventory_file playbooks/06.k8s.yaml
   ```

3. 安装过程配置

工具目录下的all.yaml为全量安装，安装效果跟依次执行playbooks目录下的01~14编号的yaml效果一致（不包括15.mindxdl.yaml）。实际安装时可根据需要对组件灵活删减

# 高级配置

## 角色介绍

本工具提供了多个ansible role。可灵活组以满足不同安装需求

### 角色：mindx.docker

安装docker-ce

### 角色：mindx.k8s.install

安装kubernetes相关二进制文件，并启动kubelet。该角色只安装，不作任何配置

### 角色：mindx.k8s.master

初始化集群。该角色将在执行的节点上执行`kubeadm init`初始化kubernetes集群，并安装calico网络插件

参数：

| 参数名                         | 说明                                                             |
| ----------------- | -------------------------------------------------------------- |
| K8S_API_SERVER_IP | 指定kubenetes的apiserver绑定的ip地址，默认空。在多网卡时建议配置，防止apiserver监听到其他网卡上 |

### 角色：mindx.k8s.worker

加入集群。该角色将在执行的节点上执行`kubeadm join`加入已经初始化好kubernetes集群。需在mindx.k8s.master之后执行
