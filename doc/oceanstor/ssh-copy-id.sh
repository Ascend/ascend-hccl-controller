#!/usr/bin/env bash

# 以下命令在172.16.0.44节点执行上

# 先手动执行 ssh-keygen 命令，生成ssh公私钥

# 以下脚本，循环执行ssh-copy-id命令，将公钥copy到远程节点172.16.0.10~172.16.0.43

for ((i=10; i<=43; i++))
do 
  ssh-copy-id 172.16.0.${i}
done
