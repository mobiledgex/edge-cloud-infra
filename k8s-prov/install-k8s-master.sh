#!/bin/sh
# must run as root
# on the master
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
sudo apt install -y python
which python
if [ $? -ne 0 ]; then
    echo python not installed
    exit 1
fi
which kubeadm
if [ $? -ne 0 ]; then
    echo missing kubeadm
    exit 1
fi
#nohup consul agent -server -bootstrap-expect=1 -data-dir=/tmp/consul -node=`hostname` -bind=$MYIP -syslog -config-dir=/etc/consul/conf.d  &
kubeadm init --apiserver-advertise-address=$MYIP --pod-network-cidr=10.244.0.0/16 --ignore-preflight-errors=all
if [ $? -ne 0 ]; then
    echo  kubeadm exited with error
    exit 1
fi
#export KUBECONFIG=/etc/kubernetes/admin.conf
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
which kubectl
if [ $? -ne 0 ]; then
    echo missing kubectl
    exit 1
fi
kubectl apply -f https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/rbac.yaml
if [ $? -ne 0 ]; then
    echo kubectl exited with error installing rbac
    exit 1
fi
#kubectl apply -f https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/canal.yaml
# use fixed version. the original will fail validation
kubectl apply -f https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/canal.yaml
if [ $? -ne 0 ]; then
    echo kubectl exited with error installing canal
    exit 1
fi
# the pod network plugin has to be done for coredns to come up
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
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
kubeadm token create --print-join-command  > /tmp/k8s-join-cmd
cat /tmp/k8s-join-cmd
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
mkdir -p /root/.kube
echo created /root/.kube
sudo cp -i /etc/kubernetes/admin.conf /root/.kube/config
echo copied admin.conf to /root/.kube/config
sudo chown root:root /root/.kube/config
cd /tmp
echo running simple http server at :8000
python -m SimpleHTTPServer 
#should not get here
echo error returned from simple http server
#consul kv put join-cmd "`cat /tmp/k8s-join-cmd`"
