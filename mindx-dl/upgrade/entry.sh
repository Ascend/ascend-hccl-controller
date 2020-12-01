#!/bin/bash

function exportYaml(){
    mkdir -p "${1}"
    cd "${1}"
    # Collect previous version resource defination of volcano.
    kubectl get deployment -n volcano-system volcano-scheduler -o yaml > volcano-scheduler-export.yaml
    kubectl get deployment -n volcano-system volcano-controllers -o yaml > volcano-controller-export.yaml
    kubectl get deployment -n volcano-system volcano-admission -o yaml > volcano-admission-export.yaml
    kubectl get job -n volcano-system volcano-admission-init -o yaml > volcano-admission-init-export.yaml || true
    # Collect previous version resource defination of device-plugin.
    kubectl get daemonset -n kube-system ascend-device-plugin-daemonset -o yaml > 910-ascend-device-plugin-export.yaml
    kubectl get daemonset -n kube-system ascend-device-plugin2-daemonset -o yaml > 310-ascend-device-plugin-export.yaml
    # Collect previous version resource defination of hccl-controller
    kubectl get deployment hccl-controller -o yaml > hccl-controller-export.yaml
    # Collect previous version resource defination of cadvisor
    kubectl get daemonset -n cadvisor cadvisor -o yaml > cadvisor-export.yaml
    cd ..
}

function printVersion(){
    echo -e "$(kubectl describe pod -n "${1}" "$(kubectl get pods -A | grep "${2}" | awk '{print $2}' | head -n 1)" | grep Image: | awk '{print $2}')"
}


function saveImageVersion(){
    vcAdmission=$(printVersion "volcano-system" "volcano-admission")
    vcControllers=$(printVersion "volcano-system" "volcano-controllers")
    vcScheduler=$(printVersion "volcano-system" "volcano-scheduler")
    hc=$(printVersion "default" "hccl-controller")
    ca=$(printVersion "cadvisor" "cadvisor")
    dp=$(printVersion "kube-system" "device-plugin")
}


function versionPrint(){
    echo -e "\nImages versions:"
    echo -e "\nVolcano:\n$vcAdmission\n$vcControllers\n$vcScheduler"
    echo -e "\nHccl-Controller:\n$hc"
    echo -e "\nCadvisor\n$ca"
    echo -e "\nAscend-device-plugin\n$dp\n"
}


function upgrade(){
    set -e
    # Save previous version image info
    saveImageVersion
    echo -e "\nBefore Upgrade:" | tee ./pre_check.txt
    versionPrint | tee ./pre_check.txt

    # Pause
    local continue
    read -r -p "Do you want to continue upgrade?(yes/no)" continue

    while [ "$continue" != 'yes' ] && [ "$continue" != 'no' ];do
          read -r -p "Invalid input. Do you want to continue upgrade?(yes/no)" continue
    done

    if [ "$continue" == 'no' ];then
        echo -e "\nUpgrade terminated."
        return 0
    fi

    # Save previous version yamls
    exportYaml "./Previous_version_info"

    # Upgrade begins
    echo -e "\nUpgrade begins.\n"
    ansible-playbook -vv ./upgrade.yaml --tags=upgrade

    # Checking upgrade result.
    echo -e "\nChecking upgrade result..."
    ansible-playbook -vv ./upgrade.yaml --tags=check | tee ./check_log.txt
    echo -e "\nChecking complete."

    # Post-upgrade processing
    if [ "$(grep -c "failed=1" ./check_log.txt)" -eq "1" ];then
        echo -e "\nUpgrade failed.\n"

        local rollback
        read -r -p "Do you want to roll back to previous version?(yes/no)" rollback

        while [ "$rollback" != 'yes' ] && [ "$rollback" != 'no' ];do
              read -r -p "Invalid input. Do you want to roll back to previous version?(yes/no)" rollback
        done

        if [ "$rollback" == 'yes' ];then
            echo -e "\nRolling back to previous version...\n"
            exportYaml "./Current_version_info"
            # Delete resource defination of current version
            cd ./Current_version_info
            kubectl delete -f 310-ascend-device-plugin-export.yaml
            kubectl delete -f 910-ascend-device-plugin-export.yaml
            kubectl delete -f cadvisor-export.yaml
            kubectl delete -f hccl-controller-export.yaml
            kubectl delete -f volcano-admission-export.yaml
            kubectl delete -f volcano-admission-init-export.yaml
            kubectl delete -f volcano-controller-export.yaml
            kubectl delete -f volcano-scheduler-export.yaml
            cd ..
            # Wait a short period of time til resources deleted.
            sleep 15s
            # Apply resource definition of previous version
            cd ./Previous_version_info
            kubectl apply -f 310-ascend-device-plugin-export.yaml
            kubectl apply -f 910-ascend-device-plugin-export.yaml
            kubectl apply -f cadvisor-export.yaml
            kubectl apply -f hccl-controller-export.yaml
            kubectl apply -f volcano-admission-export.yaml
            kubectl apply -f volcano-controller-export.yaml
            kubectl apply -f volcano-scheduler-export.yaml
            cd ..
            rm -rf ./Current_version_info
            # Print image versions
            sleep 20s
            saveImageVersion
            echo -e "\nAfter Roll back:" | tee ./post_check.txt
            versionPrint | tee ./post_check.txt
        else
            saveImageVersion
            echo -e "Roll back not applied:" | tee ./post_check.txt
            versionPrint | tee ./post_check.txt
        fi
    else
        echo "Upgrade successfully!"
        # Choose whether remove previous version images.
        local remove
        read -r -p "Do you want to remove previous version images?(yes/no)" remove
        while [ "$remove" != 'yes' ] && [ "$remove" != 'no' ];do
              read -r -p "Invalid input. Do you want to remove previous version images?(yes/no)" remove
        done
        # Remove old version images.
        if [ "$remove" == 'yes'  ];then
                ansible-playbook upgrade.yaml --tags=remove-images --extra-vars "vA=$vcAdmission vC=$vcControllers vS=$vcScheduler hc=$hc ca=$ca dp=$dp"
        fi

        rm -rf ./Previous_version_info
        # Print image versions
        saveImageVersion
        echo -e "\nAfter Upgrade:" | tee ./post_check.txt
        versionPrint | tee ./post_check.txt
    fi
    rm -f ./check_log.txt
}


function main(){
    upgrade
}

main
echo ""
echo "Finished!"
echo ""
