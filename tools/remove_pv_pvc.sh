#!/bin/bash

DL_PLATFORM_COMPONENTS="apigw,cluster,data,dataset,image,"\
"label,model,task,train,user,alarm,mysql"

function delete_components()
{
    IFS=","
    for component in ${DL_PLATFORM_COMPONENTS}
    do
        cd ~/deploy_yamls/${component}
        kubectl delete -f *.yaml
    done
    unset IFS
    kubectl delete -f ~/deploy_yamls/prometheus/prometheus.yaml
}

delete_components
