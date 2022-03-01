
# Error writing Crisocket

## 问题
[kubelet-check] Initial timeout of 40s passed.
error execution phase upload-config/kubelet: 
Error writing Crisocket information for the control-plane node: timed out waiting for the condition

## 解决方法
```bash
swapoff -a && kubeadm reset  && systemctl daemon-reload && systemctl restart kubelet  && iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X
```
当使用virtualbox虚拟机时，虚拟机默认hostname可能导致此错误，需修改hostname
