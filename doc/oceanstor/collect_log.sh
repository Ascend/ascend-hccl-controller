#! /usr/bin/env bash

# 以下命令在172.16.0.44节点执行上

readonly BASE_DIR=$(cd "$(dirname ${0})" > /dev/null 2>&1; pwd -P)

if [[ -d ${BASE_DIR}/collect_log_dir ]]; then
    rm -rf ${BASE_DIR}/collect_log_dir/*
else
    mkdir ${BASE_DIR}/collect_log_dir
fi

cd ${BASE_DIR}/collect_log_dir

kubectl get node -o wide > get_node.log

kubectl get pod -A -o wide > get_pod.log

mkdir mindx-dl_log; cd mindx-dl_log

for ((i=10; i<=43; i++)); do
    echo 172.16.0.${i}
    mkdir 172.16.0.${i}
    scp -r 172.16.0.${i}:/var/log/mindx-dl 172.16.0.${i}
done

echo 172.16.0.44; cp -r /var/log/mindx-dl 172.16.0.44

cd -

mkdir k8s_components_log; cd k8s_components_log

k8s_components=("calico-" "coredns-" "etcd-master-" "kube-apiserver-master-" "kube-controller-manager-master-" "kube-proxy-" "kube-scheduler-master-" "kube-vip-master-")

for component in ${k8s_components[*]}; do
    component_pod=$(kubectl get pods -n kube-system | grep ${component} | awk '{print $1}')
    for pod in ${component_pod}; do
        echo ${pod}
        kubectl logs --tail=2000 -n kube-system ${pod} > ${pod}.log
    done
done
