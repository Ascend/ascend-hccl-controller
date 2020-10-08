### 集群快速部署

#### 注意事项

软件包和镜像分为ARM架构和x86架构，用户需要根据实际情况获取对应的软件包和镜像包。

- 以arm64.*XXX*结尾的适用于ARM架构。
- 以amd64.*XXX*结尾的适用于x86架构。

#### 前提条件

- 已完成操作系统的安装。
- 已完成NPU驱动的安装。 
- 已完成MindX DL组件的编译。

#### 软件包

安装前需要准备所需的Ansible安装包和依赖软件，软件包列表如下。

| 用途                            | 软件名称（arm）             | 软件名称（x86）             |
| ------------------------------- | --------------------------- | --------------------------- |
| Python、Ansible离线安装包       | base-pkg_arm64.zip          | base-pkg-amd64.zip          |
| Go安装包                        | go1.14.3.linux-arm64.tar.gz | go1.14.3.linux-amd64.tar.gz |
| NFS、Docker、K8s、Git离线安装包 | offline-pkg-arm64.zip       | offline-pkg-amd64.zip       |

#### 脚本

从MindX-DL全量包中获取离线部署脚本，如下表所示。

| 脚本名称                     | 用途                                      | 全量包中的路径                    |
| ---------------------------- | ----------------------------------------- | --------------------------------- |
| entry.sh                     | 快速部署入口脚本                          | /mindx-dl/deploy/offline/cluster/ |
| set_global_env.yaml          | 设置全局变量                              | /mindx-dl/deploy/offline/cluster/ |
| offline_install_package.yaml | 安装软件包及依赖                          | /mindx-dl/deploy/offline/cluster/ |
| offline_load_images.yaml     | 导入所需Docker镜像                        | /mindx-dl/deploy/offline/cluster/ |
| init_kubernetes.yaml         | 建立K8s集群                               | /mindx-dl/deploy/offline/cluster/ |
| offline_deploy_service.yaml  | 部署MindX DL组件                          | /mindx-dl/deploy/offline/cluster/ |
| ascendplugin-volcano.yaml    | 昇腾910处理器Ascend Device Plugin配置文件 | /mindx-dl/deploy/yamls/           |
| ascendplugin-310.yaml        | 昇腾310处理器Ascend Device Plugin配置文件 | /mindx-dl/deploy/yamls/           |
| calico.yaml                  | K8s网络插件配置文件                       | /mindx-dl/deploy/yamls/           |
| hccl-controller.yaml         | NPU训练任务组件配置文件                   | /mindx-dl/deploy/yamls/           |
| gen-admission-secret.sh      | 生成Volcano组件秘钥                       | /mindx-dl/deploy/yamls/           |
| kubernetes(文件夹)           | NPU设备监控组件配置文件                   | /mindx-dl/deploy/yamls/           |
| rbac.yaml                    | K8s角色权限访问控制                       | /mindx-dl/deploy/yamls/           |
| volcano-v20.1.0.yaml         | NPU训练任务调度组件配置文件               | /mindx-dl/deploy/yamls/           |

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

1. 以**root**用户登录服务器，将Python、Ansible离线安装包拷贝到管理节点任意位置并解压，执行如下命令安装python开发环境

    ```
    dpkg -i dos2unix*.deb zlib1g-dev*.deb libffi-dev*.deb
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

2. 在解压软件包目录下执行以下命令安装Ansible

    ```
    pip3.7 install Jinja2-2.11.2* MarkupSafe-1.1.1* PyYAML-5.3.1* pycparser-2.20* cffi-1.14.3* six-1.15.0* cryptography-3.1*
    tar zxf setuptools-19.6.tar.gz
    cd setuptools-19.6
    python3.7 setup.py install
    cd ..
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

3. 执行以下命令，编辑hosts文件

    vi /etc/ansible/hosts

    根据实际写入以下内容：

    ```
    [all:vars]
    # default shared directory, you can change it as yours
    nfs_shared_dir=/data/atlas_dls
    
    # NFS service IP
    nfs_service_ip=nfs-host-ip
    
    # Master IP
    master_ip=master-host-ip
    
    # dls install package dir
    dls_root_dir=/tmp
    
    [nfs_server]
    nfs-host-name ansible_host=nfs-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    
    [master]
    master-host-name ansible_host=master-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    
    [training_node]
    training-node1-host-name ansible_host=training-node1-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    training-node2-host-name ansible_host=training-node2-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    ...
    
    [inference_node]
    inference-node1-host-name ansible_host=inference-node1-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    inference-node2-host-name ansible_host=inference-node2-host-ip ansible_ssh_user="username" ansible_ssh_pass="password"
    ...
    
    [workers:children]
    training_node
    inference_node
    
    [cluster:children]
    master
    workers
    ```

    参数说明：

    ```
    – nfs-host-ip：NFS节点服务器IP地址，根据实际写入。
    – master-host-ip：管理节点服务器IP地址，根据实际写入。
    – XXX-host-name：节点主机名，根据实际写入。
    – XXX-host-ip：节点IP地址，根据实际写入。
    – username：对应节点的用户名，根据实际写入。
    – passwd：对应节点的用户密码，根据实际写入。
    ```

4. 执行以下命令，编辑ansible.cfg

    vi /etc/ansible/ansible.cfg

    取消以下两行内容的注释并更改deprecation_warnings为“False”：

    ```
    host_key_checking = False
    deprecation_warnings = False
    ```

5. 执行部署脚本

   5.1 将软件包、镜像包和yaml文件（以x86的为例）上传至hosts中定义的dls_root_dir目录，目录结构如下（以/tmp为例）：

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
   │   ├── hccl-controller_amd64.tar.gz
   │   ├── huawei-cadvisor-beta_amd64.tar.gz
   │   ├── kube-apiserver_amd64.tar.gz
   │   ├── kube-controller-manager_amd64.tar.gz
   │   ├── kube-proxy_amd64.tar.gz
   │   ├── kube-scheduler_amd64.tar.gz
   │   ├── pause_amd64.tar.gz
   │   ├── vc-controller-manager_amd64.tar.gz
   │   ├── vc-scheduler_amd64.tar.gz
   │   └── vc-webhook-manager_amd64.tar.gz
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
    └── volcano-v20.1.0.yaml
   ```

   5.2 上传/mindx-dl/deploy/offline/cluster/下所有文件到管理节点任意目录，进入目录执行以下脚本，进行MindX DL快速部署：

   **dos2unix ***

   **bash -x entry.sh**