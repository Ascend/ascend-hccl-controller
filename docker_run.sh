#!/bin/bash


docker_name="hccl-controller"




docker stop ${docker_name}
docker rm ${docker_name}


sudo docker run --net=host --user=root:root   -it \
-v /opt/deviceplugin:/opt/deviceplugin \
--name ${docker_name}  hccl-controller:latest   /bin/bash
  
 
