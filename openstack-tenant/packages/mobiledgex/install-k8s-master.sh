#!/bin/sh
# must run as root
# on the master
set -x
if [ $# -lt 3 ]; then
	echo "Insufficient arguments"
	echo "Need interface-name master-ip my-ip"
	exit 1
fi
INTF=$1
MASTERIP=$2
MYIP=$3
echo "Interface $INTF"
echo "Master IP $MASTERIP"
echo "My IP Address: $MYIP"

systemctl is-active --quiet kubelet
if [ $? -ne 0 ]; then
  systemctl start kubelet
  systemctl enable kubelet
fi

which kubeadm
if [ $? -ne 0 ]; then
    echo missing kubeadm
    exit 1
fi
#nohup consul agent -server -bootstrap-expect=1 -data-dir=/tmp/consul -node=`hostname` -bind=$MYIP -syslog -config-dir=/etc/consul/conf.d  &
#kubeadm init --apiserver-advertise-address=$MYIP --pod-network-cidr=10.244.0.0/16 --ignore-preflight-errors=all
kubeadm init --apiserver-advertise-address=$MYIP --pod-network-cidr=192.168.0.0/16 --ignore-preflight-errors=all
if [ $? -ne 0 ]; then
    echo  kubeadm exited with error
    exit 1
fi
#export KUBECONFIG=/etc/kubernetes/admin.conf
which kubectl
if [ $? -ne 0 ]; then
    echo missing kubectl
    exit 1
fi
for d in /home/ubuntu /root; do
    mkdir -p $d/.kube
    cp  /etc/kubernetes/admin.conf $d/.kube/config
done
chown -R ubuntu:ubuntu /home/ubuntu/.kube
export KUBECONFIG=/etc/kubernetes/admin.conf
kubectl version
while [ $? -ne 0 ] ; do
    echo kubectl version failed
    sleep 7
    kubectl version
done

kubectl apply -f "/etc/mobiledgex/weave-2.8.1.yml"
if [ $? -ne 0 ] ; then
    echo Failed to install Weave
    exit 1
fi

kubectl get pods --all-namespaces
kubectl get nodes | grep NotReady
while [ $? -eq 0 ] ; do
	echo Waiting for master to be Ready
	sleep 7
	kubectl get nodes | grep NotReady
done
kubectl get nodes 
if [ $? -ne 0 ]; then
    echo kubectl exited with error doing get nodes
    exit 1
fi
kubeadm token create --ttl 0 --print-join-command | tee /tmp/k8s-join-cmd.tmp
cat /tmp/k8s-join-cmd.tmp
mv /tmp/k8s-join-cmd.tmp /var/tmp/k8s-join/k8s-join-cmd
chown ubuntu:ubuntu /var/tmp/k8s-join/k8s-join-cmd

# Start k8s-join service if not started already
systemctl is-active --quiet k8s-join
if [ $? -ne 0 ]; then
  echo "start k8s-join service"
  systemctl enable k8s-join
  systemctl start k8s-join
fi
echo master ready
