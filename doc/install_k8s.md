
# init
kubeadm init --apiserver-advertise-address=192.168.56.2 --image-repository registry.aliyuncs.com/google_containers

低版本kubeadm不支持image-repository


# get token
kubeadm token list

# get ca hash

openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt |openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex

# join
kubeadm join 192.168.56.2:6443 --token qkr3dr.h9ccuc7khlvq7mh7  --discovery-token-ca-cert-hash sha256:049f951c1cbcc8096c789edc243da176f48614c6fad15fb008fd0daf21d829ef

#资料

https://github.com/opsnull/follow-me-install-kubernetes-cluster



#dashboard

kubectl create sa dashboard-admin -n kube-system

kubectl create clusterrolebinding dashboard-admin --clusterrole=cluster-admin --serviceaccount=kube-system:dashboard-admin


kubectl get secrets -n kube-system |grep dashboard-admin

得到secret名

kubectl describe secret -n kube-system <secret-name>

获取token

