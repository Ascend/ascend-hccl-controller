# HCCL-Controller
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
![](doc/images/controller-interaction-process.png "controller-interaction-process")

<h2 id="hccl-controller.md">HCCL-Controller</h2>

HCCL-Controller is a Huawei-developed component used for NPU training tasks. It uses the Kubernetes informer mechanism to continuously monitor NPU training tasks and various events of pods, read the NPU information of pods, and generate the corresponding ConfigMap. The ConfigMap contains the  **hccl.json**  configuration file required for NPU training tasks, facilitating the NPU training tasks to better collaborate and schedule the underlying  Ascend 910 AI Processor.

## HCCL-Controller Process<a name="section2078393613277"></a>

[Figure 1](#fig13227145124720)  shows the HCCL-Controller process.

**Figure  1**  HCCL-Controller process<a name="fig13227145124720"></a>  
![](doc/images/hccl-controller-process.png "hccl-controller-process")

1.  Device-plugin periodically reports the  **DeviceID**  and health status of the  Ascend 910 AI Processor  node by using the list-and-watch API.

2.  After receiving a training job request, the scheduler creates a job and a ConfigMap. Use the Volcano scheduler to select the node where the job is to be deployed.

3.  The scheduler sends the pod creation information to the kubelet of the selected node.

4.  On the selected node, Device-plugin receives a device allocation request from kubelet and returns information such as  **DeviceID**,  **Volume**, and environment variables to the kubelet. Kubelet allocates resources to the pod.

5.  Device-plugin modifies the  **annotation**  field of the pod and writes the  Ascend 910 AI Processor  NIC IP address and the  **DeviceID**  allocated to the pod into the  **annotation**  field of the pod.

6.  The HCCL-Controller continuously monitors changes of the volcano job and pod. If a new pod is created, HCCL-Controller obtains the value of  **annotation**  from the pod. After all pod information of the volcano job is obtained, hccl-controller updates the ConfigMap of rings-config.

7.  The container training job in the pod continuously checks the status of the ConfigMap. If the status is complete, the  **hccl.json**  file can be generated based on the configmap.


## HCCL-Controller Service Rules<a name="section139091513611"></a>

HCCL-Controller is a component used to generate the  **hccl.json**  file of all pods of a training job. This component is dedicated for the  Atlas 800 training server  Kubernetes cluster.

-   <a name="li121021418717"></a>For training tasks, the  **ring-controller.atlas: ascend-910**  label needs to be set for pods and ConfigMaps. HCCL-Controller filters data using this label to distinguish the Ascend 910 scenario from other scenarios.
-   The mapping between volcano jobs and ConfigMaps is as follows: The ConfigMap name of  **volume**  \(**ascend-910-config**\) in  **volcano job.yaml**  is the ConfigMap corresponding to volcano jobs.
-   HCCL-Controller continuously monitors changes the volcano job, pod, and ConfigMap \(the changes must carry the label in  [Convention 1: Training Task, Pod, and ConfigMap](#li121021418717)\). The volcano job and ConfigMap of the same training task are associated through  **volume**  \(**ascend-910-config**\). If a new pod is created, the HCCL-Controller obtains the value of  **annotation**  \(**atlas.kubectl.kubernetes.io/ascend-910-configuration**\) in the pod and creates a data cache information table for the volcano job. After all instance information of the volcano job is obtained, the HCCL-Controller updates the ConfigMap of the corresponding rings-config.
-   The default file name of rings-config in the ConfigMap is  **hccl.json**, and the default mounting path is  **/user/serverid/devindex/config**.

## Deploying the HCCL-Controller<a name="section124015514383"></a>

1.  Run the following commands to compile the HCCL-Controller:
    ```
        cd build
    
        chmod +x build.sh
    
        ./build.sh
    ```
2.  Run the following commands to start the HCCL-Controller:
    ```
        mkdir -p /var/log/atlas_dls/hccl-controller
    
        kubectl apply -f rbac.yaml
    
        kubectl apply -f hccl-controller.yaml
    ```



<h2 id="environment-dependencies.md">Environment Dependencies</h2>

Kubernetes 1.16 or later

Go 1.13 or later

<h2 id="directory-structure.md">Directory Structure</h2>

```
hcl-controller                             # HCL-CONTROLLER module of the deep learning component
├─ build                                  # Compilation and test directory
│   ├─ build.sh
│   ├─ deploy.sh
│   ├─ Dockerfile
│   ├─ test.bat
│   └─ test.sh
├─ cmd/ring-controller
│   └─ main.go                          # Program entry
├─ doc/images                            # Document materials
├─ hack
│   ├─ update-codegen.sh
│   └─ verify-codegen.sh
├─ pkg                                  # Program file package
│   ├─ apis/resourcecontroller
│   │   ├─ v1alpha1
│   │   │   ├─ doc.go
│   │   │   ├─ register.go
│   │   │   ├─ types.go
│   │   │   └─ zz_generated.deepcopy.go
│   │   └─ register.go
│   ├─ resource-controller/signals
│   │   ├─ signal.go
│   │   ├─ signal_posix.go
│   │   └─ signal_windows.go
│   ├─ ring-controller/controller
│   │   ├─ agent_interface.go
│   │   ├─ businessagent.go
│   │   ├─ businessagent_test.go
│   │   ├─ businessworker.go
│   │   ├─ businessworker_test.go
│   │   ├─ controller.go
│   │   ├─ controller_test.go
│   │   └─ type.go
│   └─ util/waitcycle
│   │   └─ waitcycle.go
├─ docker_run.sh
├─ go.mod
├─ go.sum
└─ README.zh.md                       # README file.
```

<h2 id="version-updates.md">Version Updates</h2>

| Version   | Date   | Description  |
| ---- | ---- | ---- |
| V20.1.0| 2020-09-30    | This issue is the first official release.   |


