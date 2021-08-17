# HCCL-Controller.EN
-   [Description](#description.md)
-   [HCCL-Controller](#hccl-controller.md)
-   [Environment Dependencies](#environment-dependencies.md)
-   [Directory Structure](#directory-structure.md)
-   [Version Updates](#version-updates.md)
<h2 id="description.md">Description</h2>

-   A controller tracks at least one Kubernetes resource type. These objects have a specified field that represents the desired state. The controllers for that resource are responsible for making the current state come closer to the desired state.
-   Controller Manager is the management and control center in a cluster. It consists of multiple controllers responsible for different resources to manage all resources such as nodes and pods in the cluster.
-   Controller Manager provides the event dispatching capability. Different controllers only need to register corresponding handlers to wait for receiving and processing events.
-   Each specific resource is maintained and managed by a specific controller to retain the desired state.

**Figure  1**  Controller interaction process<a name="fig14783175555117"></a>  
![](doc/images/Controller-interaction-process.png "controller-interaction-process")

<h2 id="hccl-controller.md">HCCL-Controller</h2>

HCCL-Controller is a Huawei-developed component used for NPU training jobs. It uses the Kubernetes informer mechanism to continuously monitor NPU training jobs and various events of pods, read the NPU information of pods, and generate the corresponding ConfigMap. The ConfigMap contains the  **hccl.json**  configuration file required for NPU training jobs, enabling the NPU training jobs to better collaborate with and schedule the underlying  Ascend 910 AI Processor.

## HCCL-Controller Process<a name="section2078393613277"></a>

[Figure 1](#fig13227145124720)  shows HCCL-Controller process.

**Figure  1**  HCCL-Controller process<a name="fig13227145124720"></a>  
![](doc/images/HCCL-Controller-process.png "hccl-controller-process")

1.  Ascend Device Plugin periodically reports the  **DeviceID**  and health status of the  Ascend 910 AI Processor  node by using the list-and-watch API.

2.  After receiving a training job request, the scheduler creates a job and a ConfigMap. Use the Volcano scheduler to select the node where the job is to be deployed.

3.  The scheduler sends the pod creation information to the kubelet of the selected node.

4.  On the selected node, Ascend Device Plugin receives a device allocation request from kubelet and returns information, such as  **DeviceID**,  **Volume**, and environment variables, to kubelet. Kubelet allocates resources to the pod.

5.  Ascend Device Plugin can write the  Ascend 910 AI Processor  NIC IP address and the  **DeviceID**  allocated to the pod into the  **annotation**  field of the pod.

6.  HCCL-Controller continuously monitors changes of the volcano job and pod. If a new pod is created, HCCL-Controller obtains the value of  **annotation**  from the pod. After all pod information of the volcano job is obtained, HCCL-Controller updates the ConfigMap of rings-config.

7.  The container training job in the pod continuously checks the status of the ConfigMap. If the status is complete, the  **hccl.json**  file can be generated based on the ConfigMap.


## HCCL-Controller Service Rules<a name="section139091513611"></a>

HCCL-Controller is a component used to generate the  **hccl.json**  file of all pods of a training job. This component is dedicated for the  Atlas 800 training server  Kubernetes cluster.

-   <a name="li121021418717"></a>For training jobs, the  **ring-controller.atlas: ascend-910**  label needs to be set for pods and ConfigMaps. HCCL-Controller filters data using this label to distinguish the Ascend 910 scenario from other scenarios.
-   The mapping between volcano jobs and ConfigMaps is as follows: The ConfigMap name of  **volume**  \(**ascend-910-config**\) in  **volcano job.yaml**  is the ConfigMap corresponding to volcano jobs.
-   HCCL-Controller continuously monitors the changes of the volcano job, pod, and ConfigMap \(the changes must carry the label in  [Training Job, Pod, and ConfigMap](#li121021418717)\). The volcano job and ConfigMap of the same training job are associated through  **volume**  \(**ascend-910-config**\). If a new pod is created, HCCL-Controller obtains the value of  **annotation**  \(**ascend.kubectl.kubernetes.io/ascend-910-configuration**\) in the pod and creates a data cache information table for the volcano job. After all instance information of the volcano job is obtained, HCCL-Controller updates the ConfigMap of the corresponding rings-config.
-   The default file name of rings-config in the ConfigMap is  **hccl.json**, and the default mounting path is  **/user/serverid/devindex/config**.

## Building HCCL-Controller<a name="section124015514383"></a>

1.  Install the Go compilation environment and configure Goproxy.
2.  Run the following commands to build HCCL-Controller:

    **cd build**

    **bash build.sh**

    The files generated after building are stored in the  **output**  directory in the root directory of the source code, as shown in  [Table 1](#table1860618363516).

    **Table  1**  Files generated after building

    <a name="table1860618363516"></a>
    <table><thead align="left"><tr id="row1760620363510"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p860763675120"><a name="p860763675120"></a><a name="p860763675120"></a>File</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1860718366515"><a name="p1860718366515"></a><a name="p1860718366515"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row14578104981510"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p853441825218"><a name="p853441825218"></a><a name="p853441825218"></a>hccl-controller</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p184741133135316"><a name="p184741133135316"></a><a name="p184741133135316"></a>HCCL-Controller binary file</p>
    </td>
    </tr>
    <tr id="row1860733675117"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p13953943145215"><a name="p13953943145215"></a><a name="p13953943145215"></a>Dockerfile</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p119535431524"><a name="p119535431524"></a><a name="p119535431524"></a>Image building text file for HCCL-Controller</p>
    </td>
    </tr>
    <tr id="row11607103616516"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p12753640105215"><a name="p12753640105215"></a><a name="p12753640105215"></a>hccl-controller-<em id="i1047144135718"><a name="i1047144135718"></a><a name="i1047144135718"></a>{version}</em>.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p275310402527"><a name="p275310402527"></a><a name="p275310402527"></a>HCCL-Controller startup configuration file</p>
    </td>
    </tr>
    </tbody>
    </table>

    >![](doc/images/icon-note.gif) **NOTE:** 
    >-   _\{__version__\}_: indicates the version number. Set it based on the actual situation.
    >-   The binary dependency of ARM is different from that of x86. Therefore, compilation needs to be performed on the corresponding architecture.


## Prerequisites<a name="section2739745153910"></a>

Perform operations described in all sections except "Preparing Software Packages" in section "Preparing for Installation" in the  [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

For details, see "Installation and Deployment \> Preparations Before Installation" in the   [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

## Installing HCCL-Controller<a name="section3436132203218"></a>

For details, see "Installation and Deployment \> Installing MindX DL \> Installing HCCL-Controller" in the   [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

<h2 id="environment-dependencies.md">Environment Dependencies</h2>

-   Kubernetes 1.16 or later
-   Go 1.13 or later

<h2 id="directory-structure.md">Directory Structure</h2>

```
hccl-controller                                              # HCCL-Controller component
├── build                                                  # Build folder
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
├── pkg                                                    # Source code file
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

<h2 id="version-updates.md">Version Updates</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="en-us_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="en-us_topic_0280467800_p19856144274419"><a name="en-us_topic_0280467800_p19856144274419"></a><a name="en-us_topic_0280467800_p19856144274419"></a>Version</p>
</th>
<th class="cellrowborder" valign="top" width="33.423342334233425%" id="mcps1.1.4.1.2"><p id="en-us_topic_0280467800_p3856134219446"><a name="en-us_topic_0280467800_p3856134219446"></a><a name="en-us_topic_0280467800_p3856134219446"></a>Date</p>
</th>
<th class="cellrowborder" valign="top" width="33.24332433243324%" id="mcps1.1.4.1.3"><p id="en-us_topic_0280467800_p585634218445"><a name="en-us_topic_0280467800_p585634218445"></a><a name="en-us_topic_0280467800_p585634218445"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row5243143131115"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.2</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-07-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>Added the interaction information check with the Kubernetes.</p>
</td>
</tr>
<tr id="row533735317138"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p10908832143316"><a name="p10908832143316"></a><a name="p10908832143316"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p590810328337"><a name="p590810328337"></a><a name="p590810328337"></a>2021-03-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p1690843203317"><a name="p1690843203317"></a><a name="p1690843203317"></a>Supported the Deployment.</p>
</td>
</tr>
<tr id="row350715425123"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p162524106237"><a name="p162524106237"></a><a name="p162524106237"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="p8252121092313"><a name="p8252121092313"></a><a name="p8252121092313"></a>2020-12-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="p225281018231"><a name="p225281018231"></a><a name="p225281018231"></a>Updated the <strong id="b1397344414012"><a name="b1397344414012"></a><a name="b1397344414012"></a>Directory Structure</strong> section.</p>
</td>
</tr>
<tr id="en-us_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="en-us_topic_0280467800_p08571442174415"><a name="en-us_topic_0280467800_p08571442174415"></a><a name="en-us_topic_0280467800_p08571442174415"></a>v20.1.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.423342334233425%" headers="mcps1.1.4.1.2 "><p id="en-us_topic_0280467800_p38571542154414"><a name="en-us_topic_0280467800_p38571542154414"></a><a name="en-us_topic_0280467800_p38571542154414"></a>2020-09-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.24332433243324%" headers="mcps1.1.4.1.3 "><p id="en-us_topic_0280467800_p5857142154415"><a name="en-us_topic_0280467800_p5857142154415"></a><a name="en-us_topic_0280467800_p5857142154415"></a>This is the first official release.</p>
</td>
</tr>
</tbody>
</table>

