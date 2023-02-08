# hccl-controller
-   [组件介绍](#组件介绍)
-   [编译HCCL-Controller](#编译HCCL-Controller)
-   [组件安装](#组件安装)
-   [说明](#说明)
-   [更新日志](#更新日志)

# 组件介绍

-   一个Controller至少追踪一种类型的Kubernetes资源。这些对象有一个代表期望状态的指定字段。Controller负责确保其追踪的资源对象的当前状态接近期望状态。
-   Controller Manager就是集群内部的管理控制中心，由负责不同资源的多个Controller构成，共同负责集群内的节点、Pod等所有资源的管理。
-   Controller Manager主要提供了一个分发事件的能力，而不同的Controller只需要注册对应的Handler来等待接收和处理事件。
-   每种特定资源都有特定的Controller维护管理以保持预期状态。

**图 1**  Controller interaction process<a name="fig14783175555117"></a>  
![](doc/images/Controller-interaction-process.png "Controller-interaction-process")

## 1、HCCL-Controller整体流程<a name="section2078393613277"></a>
HCCL-Controller 是华为自研的一款用于NPU训练任务的组件，利用kubernetes的informer机制，持续监控NPU训练任务及其POD的各种事件，并读取POD的NPU信息，生成对应的
Configmap。该Configmap包含了NPU训练任务需要的hccl.json配置文件，方便NPU训练任务更好的协同和调度底层的昇腾处理器。
HCCL-Controller整体流程如[图1](#fig13227145124720)所示。

**图 1**  HCCL-Controller process<a name="fig13227145124720"></a>  
![](doc/images/HCCL-Controller-process.png "HCCL-Controller-process")

1.  Device-plugin通过list-and-watch接口，定时上报节点昇腾910处理器DeviceID和健康状态。

2.  Scheduller收到用户训练任务请求，创建Job和Configmap。使用Volacno调度器选择Job部署的节点。

3.  Scheduller发送创建Pod信息到选中的节点Kubelet上。

4.  在被选择的节点上，Device-plugin会从Kubelet收到分配设备的请求，返回DeviceID、Volume、环境变量等信息给Kubelet，Kubelet分配资源给Pod。

5.  Device-plugin修改该Pod的annotation字段，将分配给Pod的昇腾910处理器网卡IP和DeviceID写入Pod的annotation。

6.  HCCL-Controller持续监控volcano job和Pod的变化，如果有新创建的Pod，HCCL-Controller会把Pod中annotation值取出，当volcano job的所有Pod信息获取完后，更新对应rings-config的Configmap。

7.  Pod中容器训练任务持续查看Configmap的状态，发现状态为完成后，则可以从configmap中生成hccl.json文件


## 2、HCCL-Controller业务规则<a name="section139091513611"></a>

HCCL-Controller是专门用于生成训练作业所有Pod的hccl.json文件的组件，该组件为Atlas 800 训练服务器K8s集群专用组件。

-   <a name="li121021418717"></a>训练任务，Pod，ConfigMap需要设置ring-controller.atlas: ascend-910标签，HCCL-Controller通过该标签过滤，用于区分昇腾910场景和非昇腾910场景。
-   volcano job与configmap的对应方式：volcano job.yaml中volume（ascend-910-config）的configmap name，就是volcano job对应的configmap。
-   hccl-controller持续监控 volcano job，pod和ConfigMap的变化（需携带[•约定1：训练任务，Pod，ConfigMap需...](#li121021418717)中的标签），同一个训练任务的volcano job和ConfigMap通过volume（ascend-910-config）关联。如果有新创建的Pod，hccl-controller把Pod中的annotation（atlas.kubectl.kubernetes.io/ascend-910-configuration）的值取出，为volcano job创建数据缓存信息表，当volcano job的所有实例信息获取完整后，更新对应的rings-config的ConfigMap。
-   ConfigMap中rings-config的文件名默认为hccl.json，默认挂在路径为：“/user/serverid/devindex/config”。

# 编译HCCL-Controller

1.  通过git拉取源码，并切换sync-dev分支，获得ascend-hccl-controller。

    示例：源码放在/home/test/ascend-hccl-controller目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制hccl-controller、yaml文件和Dockerfile。

    **cd** _/home/test/_**ascend-hccl-controller/build/**

    **chmod +x build.sh**

    **./build.sh**

3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/test/_**ascend-hccl-controller/output**

    ```
    drwxr-xr-x 2 root root     4096 Jan 29 19:12 ./
    drwxr-xr-x 9 root root     4096 Jan 29 19:09 ../
    -r-------- 1 root root      498 Jan 29 19:09 Dockerfile
    -r-x------ 1 root root 35323904 Jan 29 19:09 hccl-controller
    -r-------- 1 root root     2374 Jan 29 19:12 hccl-controller-v3.0.0.yaml
    ```


# 组件安装

1.  请参考《MindX DL用户指南》(https://www.hiascend.com/software/mindx-dl)
    中的“集群调度用户指南 > 安装部署指导 \> 安装集群调度组件 \> 典型安装场景 \> 集群调度场景”进行。

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，如果需要加密保存，请自行修改

# 更新日志

| 版本   | 发布日期   | 修改说明  |
| ---- | ---- | ---- |
| v3.0.0| 2022-1230    | 首次发布    |

