# ceph-deploy

## 介绍
ceph-deploy是一个使用ansible安装cephfs集群的工具，它的优势是在离线环境中部署，无需访问外部网络。

本工具部署流程完全遵从ceph官方文档[链接](https://docs.ceph.com/en/pacific/)

## 软件架构

本工具支持部署Pacific（16.2.9）版本的cephfs集群

本工具支持Ubuntu 18.04 aarch64、x86_64，支持cephfs集群中只包含某种架构，或两种架构混合的场景

| 操作系统   | 版本   | CPU架构 |
|:------:|:---------:|:-------:|
| Ubuntu | 18.04     | aarch64 |
| Ubuntu | 18.04     | x86_64  |

本工具会安装如下开源软件

| 软件名      | 版本       | 备注                              |
| ---------- | ---------- | ----------------------------------|
| python3    | 3.6        | ansible会安装到python3，本机节点安装|
| ansible    | 2.11.6     | 任务编排的自动化平台，本机节点安装   |
| chrony     | 3.2        | 时间同步组件，所有节点安装          |
| docker     | 19.03.9    | 应用容器引擎，所有节点安装          |
| harbor     | 2.3.3      | 容器镜像仓，默认本机节点安装        |
| ceph       | 16.2.9     | ceph cli组件，所有节点安装         |

## 存储设备要求

Ceph拒绝在不可用的设备上配置OSD，满足以下所有条件，则认为存储设备可用：

1.  该设备必须没有分区
2.  该设备不得具有任何LVM状态
3.  不得挂载该设备
4.  该设备不得包含文件系统
5.  该设备不得包含Ceph BlueStore OSD
6.  该设备必须大于5 GB

例如如下环境，通过`lsblk`命令查询，/dev/sda存在分区、挂载点和文件系统，ceph不会在/dev/sda上配置OSD；而/dev/sdb满足以上要求，是可用的存储设备

```bash
root@ubuntu:~# lsblk
NAME       MAJ:MIN       RM        SIZE      RO      TYPE    MOUNTPOINT
sda          8:0          0        600G       0      disk
|-sda1       8:1          0        200G       0      part    /boot/efi
|-sda2       8:2          0        400G       0      part    /
sdb          8:16         0        600G       0      disk
```

ceph建议配置3个或更多节点，而且每个节点均要有可用的存储设备

请保证各个节点的系统纯净；如果节点上已安装过cephfs系统，请参考官方文档，完全清除节点上已有的ceph系统；否则，可能导致cephfs安装失败或性能下降

## 下载本工具

本工具只支持root用户，下载地址：[ceph-deploy](https://gitee.com/funnyfunny8/ceph-deploy)。2种下载方式：

1. 使用git clone

2. 下载[zip文件](https://gitee.com/funnyfunny8/ceph-deploy/repository/archive/master.zip)

然后联系工程师取得开源软件的ceph_resources.tar.gz离线安装包，将离线安装包解压在/root目录下。按如下方式放置

```bash
root@ubuntu:~# ls
ceph-deploy
ceph_resources             //由ceph_resources.tar.gz解压得到，必须放置在/root目录下
ceph_resources.tar.gz
```

## 安装教程

### 步骤1：安装ansible

工具中包含一个install_ansible.sh文件用于安装ansible

在工具目录中执行：

```bash
root@ubuntu:~/ceph-deploy# bash install_ansible.sh

root@ubuntu:~/ceph-deploy# ansible --version
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

已下步骤均在playbooks目录下进行

在playbooks/inventory文件中，需要提前规划好如下集群信息：

1. 安装harbor的服务器ip。默认为本机localhost，可更改为其他服务器ip

2. ceph_localhost节点ip，只能为本机localhost，不可更改

3. ceph_otherhost节点ip。这里至少需要配置2个或2个以上的节点ip，不可包括localhost

```ini
[harbor]
localhost ansible_connection=local

[ceph_localhost]
localhost ansible_connection=local  set_hostname="node-99"

[ceph_otherhost]
192.0.3.100  set_hostname="node-100"
192.0.3.101  set_hostname="node-101"

# 以上192.0.*.*等ip仅为示例，请修改为实际规划的ip地址
```

注意：

1. ceph要求集群内节点(ceph_localhost、ceph_otherhost）的hostname不一样，因此建议执行安装前设置所有设备使用不同的hostname。如果未统一设置且存在相同hostname的设备，那么可在inventory文件中设置set_hostname主机变量，安装过程将自动设置设备的hostname。hostname需满足ceph和ansible的格式要求，建议用“[a-z]-[0-9]”的格式，如“node-100”。例如：

### 步骤3：配置安装信息

在playbooks/group_vars目录中的all.yaml文件

```yaml
# harbor ip address
HARBOR_HOST_IP: ""
# harbor https port
HARBOR_HTTPS_PORT: 7443
# harbor install path
HARBOR_PATH: /data/harbor
# password for harbor, can not be empty, delete immediately after finished
HARBOR_PASSWORD: ""
```

其中中配置项详细为：

| 配置项               | 说明                                   |
| ----------------- | ------------------------------------ |
| HARBOR_HOST_IP    | 配置harbor的监听ip，多网卡场景下*建议配置*         |
| HARBOR_HTTPS_PORT | harbor的https监听端口，默认为7443             |
| HARBOR_PATH       | Harbor的安装路径，默认为/data/harbor                   |
| HARBOR_PASSWORD   | harbor的登录密码，不可为空，**必须配置**。**安装完成后应立即删除** |

注意：

1. harbor的登录用户名默认为admin。

### 步骤4：检查集群状态

如果playbooks/inventory_file内配置了非localhost的远程ip，根据ansible官方建议，请用户自行使用SSH密钥的方式连接到远程机器，可参考[[connection_details; Ansible Documentation](https://docs.ansible.com/ansible/latest/user_guide/connection_details.html#setting-up-ssh-keys)]

在playbooks目录中执行：

```bash
root@ubuntu:~/ceph-deploy/playbooks# ansible -i inventory_file all -m ping

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

在playbooks目录中执行：

```bash
root@ubuntu:~/ceph-deploy/playbooks# ansible-playbook -i inventory_file all.yaml
```

注：

1. ceph节点不可重复初始化，执行本步骤前，请参考官方文档，完全清除节点上已有的ceph系统

3. 如果某节点docker.service配置了代理，则可能无法访问harbor镜像仓。使用本工具前，请先在`/etc/systemd/system/docker.service.d/proxy.conf`中NO_PROXY添加harbor host的ip，然后执行`systemctl daemon-reload && systemctl restart docker`生效

4. 如果inventory_file内配置了非localhost的远程ip，本工具会将本机的/root/ceph_resources目录分发到远程机器上。如果有重复执行以上命令的需求，可在以上命令后加`-e ceph_resources_no_copy=true`参数，避免重复执行耗时的~/ceph_resources目录打包、分发操作

### 步骤6：安装后检查

检查cephfs健康状态

```bash
root@ubuntu:~# ceph -s
  cluster:
    id:     50b8227e-0d50-11ed-bb57-024216e47cd6
    health: HEALTH_OK

  services:
    mon: 3 daemons, quorum node-99,node-100,node-101 (age 3h)
    mgr: node-99.iocujm(active, since 3h), standbys: node-100.vnrgda
    mds: 1/1 daemons up, 1 standby
    osd: 5 osds: 5 up (since 3h), 5 in (since 3h)

  data:
    volumes: 1/1 healthy
    pools:   3 pools, 65 pgs
    objects: 23 objects, 8.4 kiB
    usage:   0 MiB used, 4.2 TiB / 4.2 TiB avail
    pgs:     65 active+clean
```

### 步骤7：挂载cephfs

创建cephfs的挂载目录，并手动挂载cephfs存储集群到该目录。<CEPHFS_IP>为cephfs monitor节点ip，<CEPHFS_PORT>默认为"6789"，<CEPHFS_USER>默认为"admin"，<CEPHFS_KEY>可在cephfs monitor节点通过`ceph auth get-key client.admin`命令查询

```bash
mkdir <cephfs的挂载目录>

mount -t ceph <CEPHFS_IP>:<CEPHFS_PORT>:/ <cephfs的挂载目录> -o name=<CEPHFS_USER>,secret=<CEPHFS_KEY>

```

cephfs monitor一般会部署多个节点上，建议挂载多个CEPHFS_IP，增加cephfs的高可用性，避免某个节点挂掉后挂载目录不可用；各个<CEPHFS_IP>:<CEPHFS_PORT>之间通过","分隔，最后接上":/"

```bash
mkdir <cephfs的挂载目录>

mount -t ceph <CEPHFS_IP_1>:<CEPHFS_PORT>,<CEPHFS_IP_2>:<CEPHFS_PORT>,<CEPHFS_IP_3>:<CEPHFS_PORT>:/ <cephfs的挂载目录> -o name=<CEPHFS_USER>,secret=<CEPHFS_KEY>

```

## 分步骤安装

playbooks目录下有很多文件，其中每个yaml文件对应一个组件，可以实现只安装某个组件

```bash
playbooks/
├── 01.resource.yaml  # 分发/root/ceph_resources目录
├── 02.chrony.yaml  # 安装chrony
├── 03.docker.yaml  # 安装docker
├── 04.harbor.yaml  # 安装harbor并登录
├── 05.push_image.yaml  # 推送/root/ceph_resources/images里的开源镜像到harbor
├── 06.pull_image.yaml  # 拉取部分开源镜像到各个节点
├── 07.ceph_install.yaml  # 安装ceph
├── 08.ceph_bootsrap.yaml  # 初始化ceph集群
├── 09.ceph_add_host.yaml  # 添加其他ceph节点
├── 10.cephfs_create.yaml  # 部署OSD，创建cephfs集群
```

例如:

1. 只分发软件包，则执行

   ```bash
   root@ubuntu:~/ceph-deploy/playbooks# ansible-playbook -i inventory_file 01.resource.yaml
   ```

   可在以上命令后加`-e ceph_resources_no_copy=true`参数，该参数作用请见<a href="#resources_no_copy">步骤5：执行安装注意事项第3点</a>

2. 只初始化ceph集群，则执行

   ```bash
   root@ubuntu:~/ceph-deploy/playbooks# ansible-playbook -i inventory_file 08.ceph_bootsrap.yaml
   ```

   ceph节点不可重复初始化，执行本步骤前，请参考官方文档，完全清除节点上已有的ceph系统

   由于ansible的幂等性，除08.ceph_bootsrap.yaml步骤外，其他步骤均可以重复执行

3. 工具目录下的all.yaml为全量安装，安装效果跟依次执行01~10编号的yaml效果一致。实际安装时可根据需要对组件灵活删减
