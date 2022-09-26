#!/usr/bin/expect

# 以下命令在172.16.0.44节点执行上

# 先手动执行 ssh-keygen 命令，交互式输入回车，即可默认生成ssh公私钥

# 然后用chmod u+x ssh-copy-id.sh，执行./ssh-copy-id.sh，即可循环执行ssh-copy-id命令，将公钥copy到远程节点172.16.0.10~172.16.0.43

set timeout 5

for {set i 10} {${i}<=43} {incr i} {
  spawn ssh-copy-id 172.16.0.${i}
  expect {
    "(yes/no)?" { send "yes\n"; exp_continue }
    "password"  { send "Huawei12#$\n" }
  }
#  expect eof
}
