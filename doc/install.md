# 快速入门

## 功能描述

本文主要介绍如何使用ansible安装mindxdl所需开源软件安装。其中包含如下开源软件

| 软件名        | 备注                    |
| ---------- | --------------------- |
| docker     | 集群中所有节点都需要安装          |
| kubernetes | k8s平台                 |
| mysql      | 安装在k8s集群中，挂载host的文件系统 |
| nfs        | 所有节点都需要安装nfsclient    |
| harbor     | 容器镜像仓                 |
| prometheus | 安装在kubernetes集群中      |
| grafana    | 安装在kubernetes集群中      |

## 环境要求

### 支持的操作系统说明

| 操作系统   | 版本        | CPU架构   |
|:------:|:---------:|:-------:|
| Ubuntu | 18.04.1/5 | aarch64 |
| Ubuntu | 18.04.1/5 | x86_64  |

### 支持的硬件形态说明

| 中心推理硬件         | 中心训练硬件     |
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

ansible默认安装在系统自带python3中，安装完成后执行ansible --version查看ansible是否安装成功

### 步骤2：配置集群信息

在inventory文件中，需要提前规划好如下集群信息：

1. 安装harbor的服务器ip

2. master节点ip，只能为本机localhost

3. work节点ip（默认无，请根据需要添加，不可包括master节点，即不可包括localhost）

4. mysql安装的节点ip，只能为本机localhost

5. nfs服务器ip。nfs可使用已有nfs服务器。

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

# mindx k8s namespace
K8S_NAMESPACE: "mindx-dl"
# ip address for api-server
K8S_API_SERVER_IP: ""

# nfs shared path, can be multiple configurations
NFS_PATH: ["/data/atlas_dls"]

# mysql install path
MYSQL_DATAPATH: /data/mysql
# password for mysql, can not be empty, delete immediately after finished
MYSQL_PASSWORD: ""

#mindx user
MINDX_USER: hwMindX
MINDX_USER_ID: 9000
MINDX_GROUP: hwMindX
MINDX_GROUP_ID: 9000
```

其中中配置项详细为：

| 配置项               | 说明                                   |
| ----------------- | ------------------------------------ |
| HARBOR_HOST_IP    | 配置harbor的监听ip，多网卡场景下**建议配置**         |
| HARBOR_HTTPS_PORT | harbor的https监听端口，默认为7443             |
| HARBOR_PATH       | Harbor的安装路径                          |
| HARBOR_PASSWORD   | harbor的登录密码，不可为空，**必须配置**。安装完成后应立即删除 |
| K8S_NAMESPACE     | mindx dl组件默认k8s命名空间                  |
| K8S_API_SERVER_IP | K8s的api server监听地址，多网卡场景下**建议配置**    |
| NFS_PATH          | nfs服务器的共享路径，可配置多个路径                  |
| MYSQL_DATAPATH    | mysql的安装路径                           |
| MYSQL_PASSWORD    | mysql的登录密码，不可为空，**必须配置**。安装完成后应立即删除  |
| MINDX_USER        | mindx dl组件默认运行用户                     |
| MINDX_USER_ID     | mindx dl组件默认运行用户id                   |
| MINDX_GROUP       | mindx dl组件默认运行用户组                    |
| MINDX_GROUP_ID    | mindx dl组件默认运行用户组id                  |

harbor的登录用户名默认为admin

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

1. k8s节点不可重复初始化或加入，使用本工具前，请先执行`kubeadm reset`清除节点上已有的k8s配置

2. 如果docker.service配置了代理，则可能无法访问harbor镜像仓。使用本工具前，请先在`/etc/systemd/system/docker.service.d/proxy.conf`中NO_PROXY添加harbor host的ip，然后执行`systemctl daemon-reload && systemctl restart docker`生效

3. 如果inventory_file内配置了非localhost的远程ip，本工具会将本机的~/resources目录分发到远程机器上。如果有重复执行以上命令的需求，可在以上命令后加`-e resources_no_copy=true`参数，避免重复执行耗时的~/resources目录打包、分发操作。

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
default       node-exporter-ds5f5                        1/1     Running   0          19h
default       node-exporter-s5j9s                        1/1     Running   1          19h
kube-system   calico-kube-controllers-659bd7879c-l7q55   1/1     Running   2          19h
kube-system   calico-node-5zk76                          1/1     Running   1          19h
kube-system   calico-node-cxhdn                          1/1     Running   0          19h
kube-system   coredns-f9fd979d6-l42rb                    1/1     Running   2          19h
kube-system   coredns-f9fd979d6-x2bg2                    1/1     Running   2          19h
kube-system   etcd-node-10-0-2-15                        1/1     Running   1          19h
kube-system   grafana-core-58664d599b-4d8s8              1/1     Running   1          19h
kube-system   kube-apiserver-node-10-0-2-15              1/1     Running   1          19h
kube-system   kube-controller-manager-node-10-0-2-15     1/1     Running   5          19h
kube-system   kube-proxy-g65rn                           1/1     Running   1          19h
kube-system   kube-proxy-vqzb7                           1/1     Running   0          19h
kube-system   kube-scheduler-node-10-0-2-15              1/1     Running   4          19h
kube-system   prometheus-577fb6b799-k6mwl                1/1     Running   1          19h
mindx-dl      mysql-55569fc484-bb6kw                     1/1     Running   1          19h
```

注：

1. 手动执行kubectl命令时，需取消http(s)_proxy网络代理配置，否则会连接报错或一直卡死

### 步骤7：安装MindX DL组件

1. 在~/resources/目录下创建mindxdl目录。如果该目录已存在，请确保该目录下为空
   
   ```bash
      mkdir -p ~/resources/mindxdl
   ```

2. 将MindX DL组件放到~/resource/mindxdl目录中
   
   ```bash
   ~/resources/
    `── mindxdl
        ├── Ascend-mindxdl-volcano_{version}-{arch}.zip
        ├── Ascend-mindxdl-hccl-controller_{version}-{arch}.zip
         ....
   ```

3. 在工具目录中执行安装命令

   ```bash
   root@master:~/mindxdl-deployer# ansible-playbook -i inventory_file playbooks/11.mindxdl.yaml
   ```

注：

1. MindX DL相关组件安装时依赖harbor。安装过程会制作镜像并上传到harbor中

2. 安装MindX DL组件，当前仅支持k8s为master单机节点，或worker与master节点的CPU架构相同的情况

## 更新MindX DL组件

如果用户已完整执行过以上安装步骤，本工具支持单独更新MindX DL组件。

1. 查阅“步骤2：配置集群信息”的inventory文件和“步骤3：配置安装信息”的group_vars/all.yaml文件，确保这2个配置文件同上一次使用本工具时的配置完全一致

2. 执行“步骤7：安装MindX DL组件”。该步骤可重复执行

# 详细说明

## 分步骤安装

playbooks目录下有很多文件，其中每个yaml文件对应一个组件，可以实现只安装某个组件

```bash
playbooks/
├── 01.resource.yaml
├── 02.docker.yaml
├── 03.harbor.yaml
├── 04.k8s.yaml
├── 05.mysql.yaml
├── 06.nfs.yaml
├── 07.prometheus.yaml
├── 08.kubeedge.yaml
├── 09.pre-image.yaml
├── 10.redis.yaml
├── 11.mindxdl.yaml
```

例如:

1. 分发软件包
   
   ```bash
   ansible-playbook -i inventory_file playbooks/01.resource.yaml
   ```

   可在以上命令后加`-e resources_no_copy=true`参数，该参数作用请见<a href="#resources_no_copy">步骤5：执行安装注意事项第3点</a>

2. 只安装docker，则执行
   
   ```bash
   ansible-playbook -i inventory_file playbooks/02.docker.yaml
   ```

## 安装过程配置

工具目录下的all.yaml为全量安装，安装效果跟依次执行playbooks目录下的01~10编号的yaml效果一致。实际安装时可根据需要对组件灵活删减

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
| --------------------------- | -------------------------------------------------------------- |
| apiserver_advertise_address | 指定kubenetes的apiserver绑定的ip地址，默认空。在多网卡时建议配置，防止apiserver监听到其他网卡上 |

### 角色：mindx.k8s.worker

加入集群。该角色将在执行的节点上执行`kubeadm join`加入已经初始化好kubernetes集群。需在mindx.k8s.master之后执行
