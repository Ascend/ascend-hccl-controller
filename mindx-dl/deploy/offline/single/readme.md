### 单机快速部署

#### 注意事项

软件包和镜像分为ARM架构和x86架构，用户需要根据实际情况获取对应的软件包和镜像包。

- 以arm64.*XXX*结尾的适用于ARM架构。
- 以amd64.*XXX*结尾的适用于x86架构。

#### 前提条件

- 已完成操作系统的安装。
- 已完成NPU驱动的安装。

#### 软件包

安装前需要准备所需的Ansible安装包和依赖软件，软件包列表如下。

| 用途                            | 软件名称（arm）             | 软件名称（x86）             |
| ------------------------------- | --------------------------- | --------------------------- |
| Python、Ansible离线安装包       | base-pkg_arm64.zip          | base-pkg-amd64.zip          |
| Go安装包                        | go1.14.3.linux-arm64.tar.gz | go1.14.3.linux-amd64.tar.gz |
| NFS、Docker、K8s、Git离线安装包 | offline-pkg-arm64.zip       | offline-pkg-amd64.zip       |

#### 脚本

从MindX-DL全量包中获取离线部署脚本，如下表所示。

| 脚本名称                     | 用途                       | 全量包中的路径                                               |
| ---------------------------- | -------------------------- | ------------------------------------------------------------ |
| entry.sh                     | 快速部署入口脚本           | /deploy/playbooks/offline/single/                            |
| set_global_env.yaml          | 设置全局变量               | /deploy/playbooks/offline/single/                            |
| offline_install_package.yaml | 安装软件包及依赖           | /deploy/playbooks/offline/single/                            |
| offline_load_images.yaml     | 导入所需Docker镜像         | /deploy/playbooks/offline/single/                            |
| init_kubernetes.yaml         | 建立K8s集群                | /deploy/playbooks/offline/single/                            |
| offline_deploy_service.yaml  | 部署MindX DL组件           | /deploy/playbooks/offline/single/                            |
| authority.yaml               | MindX DL服务访问数据库认证 | /src/apigw/configuration/                                    |
| mysql                        | 部署时重新生成MySQL镜像    | /src/mysql下所有脚本                                         |
| MindX DL                     | 部署服务配置文件           | MindX DL/deploy/yaml下所有脚本                               |
| MindX DL-core                | 部署服务配置文件           | MindX DL-core-device-plugin：ascend-device-plugin/ascendplugin.yamlMindX DL-core-cadvisor：/deploy/kubernetes/下所有脚本MindX DL-core-volcano：/volcano/volcano-v0.0.1.yaml |

#### 镜像包
全量镜像包列表，如下表所示。

| 镜像列表                            | 镜像包名称（arm）                                        | 镜像包名称（x86）                              |
| ------------------------------- | ------------------------------------------------------- | ------------------------------- |
| MindX DL K8s设备插件      | Ascend-K8sDevicePlugin-0.0.1-arm64-Docker.tar.gz                   | Ascend-K8sDevicePlugin-0.0.1-x86-Docker.tar.gz |
| K8s网络插件 | calico-cni_arm64.tar.gz | calico-cni_amd64.tar.gz |
|  | calico-kube-controllers_arm64.tar.gz | calico-kube-controllers_amd64.tar.gz |
|  | calico-node_arm64.tar.gz | calico-node_amd64.tar.gz |
|  | calico-pod2daemon-flexvol_arm64.tar.gz | calico-pod2daemon-flexvol_amd64.tar.gz |
| K8s域名服务 | coredns_arm64.tar.gz | coredns_amd64.tar.gz |
| K8s集群数据库 | etcd_arm64.tar.gz | etcd_amd64.tar.gz |
| MindX DL-core训练任务集合通信插件 | hccl-controller.tar.gz | hccl-controller.tar.gz |
| MindX DL-core设备监控插件 | huawei-cadvisor-beta_arm64.tar.gz | huawei-cadvisor-beta_amd64.tar.gz |
| K8s集群数据中心 | kube-apiserver_arm64.tar.gz | kube-apiserver_amd64.tar.gz |
| K8s集群管理控制器 | kube-controller-manager_arm64.tar.gz | kube-controller-manager_amd64.tar.gz |
| K8s集群通信与负载均衡 | kube-proxy_arm64.tar.gz | kube-proxy_amd64.tar.gz |
| K8s集群调度器 | kube-scheduler_arm64.tar.gz | kube-scheduler_amd64.tar.gz |
| K8s基础容器 | pause_arm64.tar.gz | pause_amd64.tar.gz |
| MindX DL-core任务调度插件 | vc-controller-manager.tar.gz | vc-controller-manager.tar.gz |
|  | vc-scheduler.tar.gz | vc-scheduler.tar.gz |
|  | vc-webhook-manager.tar.gz | vc-webhook-manager.tar.gz |



1. 将Python、Ansible离线安装包拷贝到管理节点任意位置并解压，执行如下命令安装python开发环境

```
tar -zxvf Python-3.7.5.tgz
cd Python-3.7.5
./configure --prefix=/usr/local/python3.7.5 --enable-shared
make
sudo make install
sudo cp /usr/local/python3.7.5/lib/libpython3.7m.so.1.0 /usr/lib
sudo ln -s /usr/local/python3.7.5/bin/python3 /usr/bin/python3.7
sudo ln -s /usr/local/python3.7.5/bin/pip3 /usr/bin/pip3.7
sudo ln -s /usr/local/python3.7.5/bin/python3 /usr/bin/python3.7.5
sudo ln -s /usr/local/python3.7.5/bin/pip3 /usr/bin/pip3.7.5
```

2、在解压软件包目录下执行以下命令安装Ansible

```
dpkg -i libhavege1_1.9.1-6*.deb
dpkg -i haveged_1.9.1-6*.deb
tar -zxvf ansible-2.9.7.tar.gz
cd ansible-2.9.7
python3.7 setup.py install --record files.txt
mkdir -p /etc/ansible
cp -rf examples/ansible.cfg examples/hosts /etc/ansible/
ln -sf /usr/local/python3.7.5/bin/ansible* /usr/local/bin/
cd ..
dpkg -i sshpass_1.06-1*.deb
```

3、执行以下命令，编辑hosts文件

vi /etc/ansible/hosts

根据实际写入以下内容：

```
[all:vars]
# default shared directory, you can change it as yours
nfs_shared_dir=/data/atlas_dls

# NFS service IP
nfs_service_ip={IP}

# dls install package dir
dls_root_dir=/tmp

[local]
localnode ansible_host={IP} ansible_ssh_user="{username}" ansible_ssh_pass="{passwd}"
```

参数说明：

{IP}：Atlas 800 训练服务器IP地址。

{username}：登录Atlas 800 训练服务器的用户名。

{passwd}：登录Atlas 800 训练服务器的用户密码。




```
执行以下命令，编辑ansible.cfg

vi /etc/ansible/ansible.cfg

取消以下两行内容的注释并更改deprecation_warnings为“False”：
host_key_checking = False
deprecation_warnings = False
```

4、执行部署脚本

部署脚本内容：

| playbook                                      | 用途                   |
| --------------------------------------------- | ---------------------- |
| ansible-playbook set_global_env.yaml          | 设置全局变量           |
| ansible-playbook offline_install_package.yaml | 离线安装软件包及依赖   |
| ansible-playbook offline_load_images.yaml     | 离线导入所需docker镜像 |
| ansible-playbook init_kubernetes.yaml         | 建立k8s集群            |
| ansible-playbook offline_deploy_service.yaml  | 部署MindX DL组件       |

将软件包及镜像包（以x86的为例）上传至hosts中定义的dls_root_dir目录，在管理节点上执行以下脚本：

**bash -x entry.sh**

dls_root_dir目录结构如下（以/tmp为例）：

```
/tmp
├─ docker_images
│   ├── Ascend-K8sDevicePlugin-0.0.1-x86-Docker.tar.gz
│   ├── calico-cni_amd64.tar.gz
│   ├── calico-kube-controllers_amd64.tar.gz
│   ├── calico-node_amd64.tar.gz
│   ├── calico-pod2daemon-flexvol_amd64.tar.gz
│   ├── coredns_amd64.tar.gz
│   ├── etcd_amd64.tar.gz
│   ├── hccl-controller.tar.gz
│   ├── huawei-cadvisor-beta.tar.gz
│   ├── kube-apiserver_amd64.tar.gz
│   ├── kube-controller-manager_amd64.tar.gz
│   ├── kube-proxy_amd64.tar.gz
│   ├── kube-scheduler_amd64.tar.gz
│   ├── pause_amd64.tar.gz
│   ├── vc-controller-manager.tar.gz
│   ├── vc-scheduler.tar.gz
│   └── vc-webhook-manager.tar.gz
├─ go1.14.3.linux-amd64.tar.gz
├─ offline-pkg-amd64.zip
└─ yaml
    ├── ascendplugin-volcano.yaml
    ├── ascendplugin-310.yaml
    ├── calico.yaml
    ├── hccl-controller.yaml
    ├── kubernetes
    │   ├── base
    │   │   ├── cluserrolebinding.yaml
    │   │   ├── cluserrole.yaml
    │   │   ├── daemonset.yaml
    │   │   ├── kustomization.yaml
    │   │   ├── namespace.yaml
    │   │   ├── podsecuritypolicy.yaml
    │   │   └── serviceaccount.yaml
    │   ├── overlays
    │   │   ├── examples
    │   │   │   ├── cadvisor-args.yaml
    │   │   │   ├── critical-priority.yaml
    │   │   │   ├── gpu-privilages.yaml
    │   │   │   ├── kustomization.yaml
    │   │   │   └── stackdriver-sidecar.yaml
    │   │   └── huawei
    │   │       ├── cadvisor-args.yaml
    │   │       ├── gpu-privilages.yaml
    │   │       └── kustomization.yaml
    │   └── README.md
    └── volcano-v0.0.1.yaml
```



