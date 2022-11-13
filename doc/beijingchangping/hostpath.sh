#!/usr/bin/env bash

# 以下命令在172.16.0.44节点执行如下命令，分发这个脚本到所有节点执行
# ansible all -m script -a "hostpath.sh"

# 事前准备：已安装好oceanstore的dpc客户端（由oceanstore存储完成）；创建oceanstore的挂载目录，并手动挂载oceanstore存储集群到该目录。。。
mkdir /dl  # 创建oceanstore的挂载目录
chown 9000:9000 /dl  # 使用hostpath方式，需要将挂载目录属主设置为9000
mount -t dpc <oceanstore标识符> /dl  # <oceanstore标识符>，由oceanstore存储提供，不可与“/dl”同名
# 使用autofs设置dpc开机自动挂载，以达到高可用  # 具体操作由oceanstore存储提供。。。


# 然后在172.16.0.44节点，执行一个脚本即可，tools/create_storage_dir.sh
