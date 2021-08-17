# HCCL-Controller.ZH
-   [Controller介绍](#Controller介绍.md)
-   [HCCL-Controller](#HCCL-Controller.md)
-   [环境依赖](#环境依赖.md)
-   [目录结构](#目录结构.md)
-   [版本更新信息](#版本更新信息.md)
<h2 id="Controller介绍.md">Controller介绍</h2>

-   一个Controller至少追踪一种类型的Kubernetes资源。这些对象有一个代表期望状态的指定字段。Controller负责确保其追踪的资源对象的当前状态接近期望状态。
-   Controller Manager就是集群内部的管理控制中心，由负责不同资源的多个Controller构成，共同负责集群内的节点、Pod等所有资源的管理。
-   Controller Manager主要提供了一个分发事件的能力，而不同的Controller只需要注册对应的Handler来等待接收和处理事件。
-   每种特定资源都有特定的Controller维护管理以保持预期状态。

**图 1**  Controller interaction process<a name="fig14783175555117"></a>  
![](doc/images/Controller-interaction-process.png "Controller-interaction-process")

<h2 id="HCCL-Controller.md">HCCL-Controller</h2>

HCCL-Controller是华为研发的一款用于NPU训练任务的组件，利用Kubernetes的informer机制，持续监控NPU训练任务及其Pod的各种事件，并读取Pod的NPU信息，生成对应的Configmap。该Configmap包含了NPU训练任务需要的hccl.json配置文件，方便NPU训练任务更好的协同和调度底层的昇腾910 AI处理器。

## HCCL-Controller整体流程<a name="section2078393613277"></a>

HCCL-Controller整体流程如[图1](#fig13227145124720)所示。

**图 1**  HCCL-Controller process<a name="fig13227145124720"></a>  
![](doc/images/HCCL-Controller-process.png "HCCL-Controller-process")

1.  Ascend Device Plugin通过list-and-watch接口，定时上报节点昇腾910 AI处理器DeviceID和健康状态。

2.  Scheduler收到用户训练任务请求，创建Job和Configmap。使用Volacno调度器选择Job部署的节点。

3.  Scheduler发送创建Pod信息到选中的节点Kubelet上。

4.  在被选择的节点上，Ascend Device Plugin会从Kubelet收到分配设备的请求，返回DeviceID、Volume、环境变量等信息给Kubelet，Kubelet分配资源给Pod。

5.  Ascend Device Plugin修改该Pod的annotation字段，将分配给Pod的昇腾910 AI处理器网卡IP和DeviceID写入Pod的annotation。

6.  HCCL-Controller持续监控volcano job和Pod的变化，如果有新创建的Pod，HCCL-Controller会把Pod中annotation值取出，当volcano job的所有Pod信息获取完后，更新对应rings-config的Configmap。

7.  Pod中容器训练任务持续查看Configmap的状态，发现状态为完成后，则可以从configmap中生成hccl.json文件。


## HCCL-Controller业务规则<a name="section139091513611"></a>

HCCL-Controller是专门用于生成训练作业所有Pod的hccl.json文件的组件，该组件为Atlas 800 训练服务器K8s集群专用组件。

-   <a name="li121021418717"></a>训练任务，Pod，ConfigMap需要设置ring-controller.atlas: ascend-910标签，HCCL-Controller通过该标签过滤，用于区分昇腾910场景和非昇腾910场景。
-   volcano job与configmap的对应方式：volcano job.yaml中volume（ascend-910-config）的configmap name，就是volcano job对应的configmap。
-   HCCL-Controller持续监控volcano job，pod和ConfigMap的变化（需携带[训练任务，Pod，ConfigMap](#li121021418717)中的标签），同一个训练任务的volcano job和ConfigMap通过volume（ascend-910-config）关联。如果有新创建的Pod，HCCL-Controller把Pod中的annotation（ascend.kubectl.kubernetes.io/ascend-910-configuration）的值取出，为volcano job创建数据缓存信息表，当volcano job的所有实例信息获取完整后，更新对应的rings-config的ConfigMap。
-   ConfigMap中rings-config的文件名默认为hccl.json，默认挂在路径为：“/user/serverid/devindex/config”。

## 编译HCCL-Controller<a name="section124015514383"></a>

1.  安装Go的编译环境和goporxy的配置。
2.  执行以下命令，编译HCCL-Controller。

    **cd build**

    **bash build.sh**

    编译生成的文件在源码根目录下的“output“目录，如[表1](#table1860618363516)所示。

    **表 1**  编译生成的文件列表

    <a name="table1860618363516"></a>
    <table><thead align="left"><tr id="row1760620363510"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p860763675120"><a name="p860763675120"></a><a name="p860763675120"></a>文件名</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1860718366515"><a name="p1860718366515"></a><a name="p1860718366515"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row14578104981510"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p853441825218"><a name="p853441825218"></a><a name="p853441825218"></a>hccl-controller</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p184741133135316"><a name="p184741133135316"></a><a name="p184741133135316"></a>HCCL-Controller二进制文件</p>
    </td>
    </tr>
    <tr id="row1860733675117"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p13953943145215"><a name="p13953943145215"></a><a name="p13953943145215"></a>Dockerfile</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p119535431524"><a name="p119535431524"></a><a name="p119535431524"></a>HCCL-Controller镜像构建文本文件</p>
    </td>
    </tr>
    <tr id="row11607103616516"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p12753640105215"><a name="p12753640105215"></a><a name="p12753640105215"></a>hccl-controller-<em id="i1047144135718"><a name="i1047144135718"></a><a name="i1047144135718"></a>{version}</em>.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p275310402527"><a name="p275310402527"></a><a name="p275310402527"></a>HCCL-Controller的启动配置文件</p>
    </td>
    </tr>
    </tbody>
    </table>

    >![](doc/images/icon-note.gif) **说明：** 
    >-   _\{__version__\}_：表示版本号，请根据实际写入。
    >-   arm和x86的二进制依赖不同，需要在对应架构上进行编译。


## 安装前准备<a name="section2739745153910"></a>

需要先完成《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》“安装前准备”章节中除“准备软件包”章节之外的其他章节内容。

请参考《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》中的“安装部署 \> 安装前准备”。

## 安装HCCL-Controller<a name="section3436132203218"></a>

请参考《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》中的“安装部署 \> 安装MindX DL \> 安装HCCL-Controller”。

<h2 id="环境依赖.md">环境依赖</h2>

-   Kubernetes 1.16及以上
-   Go 1.13及以上

<h2 id="目录结构.md">目录结构</h2>

```
hccl-controller                                              #hccl-controller 组件
├── build                                                  #编译构建文件夹
│   ├── build.sh
│   ├── Dockerfile
│   ├── hccl-controller.yaml
│   ├── rbac.yaml
│   └── test.sh
├── doc
│   └── images
│       ├── Controller-interaction-process.png
│       ├── HCCL-Controller-process.png
│       ├── icon-caution.gif
│       ├── icon-danger.gif
│       ├── icon-note.gif
│       ├── icon-notice.gif
│       ├── icon-tip.gif
│       └── icon-warning.gif
├── go.mod
├── go.sum
├── main.go
├── output
├── pkg                                                    #源码文件
│   ├── hwlog
│   │   └── logger.go
│   ├── resource-controller
│   │   └── signals
│   │       ├── signal.go
│   │       ├── signal_posix.go
│   │       └── signal_windows.go
│   └── ring-controller
│       ├── agent
│       │   ├── businessagent.go
│       │   ├── businessagent_test.go
│       │   ├── deploymentworker.go
│       │   ├── deploymentworker_test.go
│       │   ├── types.go
│       │   ├── vcjobworker.go
│       │   └── vcjobworker_test.go
│       ├── controller
│       │   ├── controller.go
│       │   ├── controller_test.go
│       │   └── types.go
│       ├── model
│       │   ├── deployment.go
│       │   ├── deployment_test.go
│       │   ├── types.go
│       │   ├── vcjob.go
│       │   └── vcjob_test.go
│       └── ranktable
│           ├── v1
│           │   ├── ranktable.go
│           │   ├── ranktable_test.go
│           │   └── types.go
│           └── v2
│               ├── ranktable.go
│               ├── ranktable_test.go
│               └── types.go
├── README_EN.md
└── README.md
```

<h2 id="版本更新信息.md">版本更新信息</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0280467800_p19856144274419"><a name="zh-cn_topic_0280467800_p19856144274419"></a><a name="zh-cn_topic_0280467800_p19856144274419"></a>版本</p>
</th>
<th class="cellrowborder" valign="top" width="33.423342334233425%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0280467800_p3856134219446"><a name="zh-cn_topic_0280467800_p3856134219446"></a><a name="zh-cn_topic_0280467800_p3856134219446"></a>发布日期</p>
</th>
<th class="cellrowborder" valign="top" width="33.24332433243324%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0280467800_p585634218445"><a name="zh-cn_topic_0280467800_p585634218445"></a><a name="zh-cn_topic_0280467800_p585634218445"></a>修改说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row5243143131115"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.2</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-07-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>增加和K8s交互信息的检查。</p>
</td>
</tr>
<tr id="row533735317138"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p10908832143316"><a name="p10908832143316"></a><a name="p10908832143316"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p590810328337"><a name="p590810328337"></a><a name="p590810328337"></a>2021-03-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p1690843203317"><a name="p1690843203317"></a><a name="p1690843203317"></a>支持Deployment工作负载。</p>
</td>
</tr>
<tr id="row350715425123"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p162524106237"><a name="p162524106237"></a><a name="p162524106237"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p8252121092313"><a name="p8252121092313"></a><a name="p8252121092313"></a>2020-12-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p225281018231"><a name="p225281018231"></a><a name="p225281018231"></a>更新目录结构章节。</p>
</td>
</tr>
<tr id="zh-cn_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0280467800_p08571442174415"><a name="zh-cn_topic_0280467800_p08571442174415"></a><a name="zh-cn_topic_0280467800_p08571442174415"></a>v20.1.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0280467800_p38571542154414"><a name="zh-cn_topic_0280467800_p38571542154414"></a><a name="zh-cn_topic_0280467800_p38571542154414"></a>2020-09-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0280467800_p5857142154415"><a name="zh-cn_topic_0280467800_p5857142154415"></a><a name="zh-cn_topic_0280467800_p5857142154415"></a>第一次正式发布。</p>
</td>
</tr>
</tbody>
</table>

