#!/bin/sh
# must run as root
# on both master and nodes
set -x
swapoff -a
apt-get update && apt-get install -y apt-transport-https curl unzip python
apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
#curl -fsSL https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/docker-gpg | apt-key add -
add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
apt-get update && apt-get install -y docker-ce
which docker
if [ $? -ne 0 ]; then
    echo docker install failed
    exit 1
fi
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
##curl -s https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/google-apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
#apt-get update && apt-get install -y kubelet=1.12.4-00 kubeadm=1.12.4-00 kubectl=1.12.4-00
# list of kubernetes versions are available at
#    curl -s https://packages.cloud.google.com/apt/dists/kubernetes-xenial/main/binary-amd64/Packages | grep Version | awk '{print $2}'

#apt-get update && apt-get install -y kubelet=1.11.2-00 kubeadm =1.11.2-00kubectl=1.11.2-00
apt-get update && apt-get install -y kubelet kubeadm kubectl
#curl -o /usr/bin/kubectl -s -LO https://storage.googleapis.com/kubernetes-release/release/v1.12.1/bin/linux/amd64/kubectl
#curl -o /usr/bin/kubeadm -s -LO https://storage.googleapis.com/kubernetes-release/release/v1.12.1/bin/linux/amd64/kubeadm
#curl -o /usr/bin/kubelet -s -LO https://storage.googleapis.com/kubernetes-release/release/v1.12.1/bin/linux/amd64/kubelet
chmod a+rx /usr/bin/kubeadm /usr/bin/kubelet /usr/bin/kubectl
which kubectl
if [ $? -ne 0 ]; then
    echo kubectl not installed
    exit 1
fi
which kubeadm
if [ $? -ne 0 ]; then
    echo kubeadm not installed
    exit 1
fi
#curl -s -LO https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/kubelet.config.yaml
## v1.12.1  is looking for this config.yaml
#cp kubelet.config.yaml /var/lib/kubelet/config.yaml
sed -i "s/cgroup-driver=systemd/cgroup-driver=cgroupfs/g" /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
systemctl restart kubelet
#wget --quiet https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz
#tar xf go1.10.2.linux-amd64.tar.gz 
#export PATH=`pwd`/go/bin:$PATH
#export GOPATH=/usr/local
#which go
#if [ $? -ne 0 ]; then
#    echo go not installed
#    exit 1
#fi
#go get github.com/kubernetes-incubator/cri-tools/cmd/crictl
curl https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/crictl -o /usr/local/bin/crictl
chmod +x /usr/local/bin/crictl
which crictl
if [ $? -ne 0 ]; then
    echo crictl not installed
    exit 1
fi
kubeadm config images pull
echo install-k8s-base.sh ok
#mkdir -p /var/lib/consul
#mkdir -p /usr/share/consul
#mkdir -p /etc/consul/conf.d
#curl -OL https://releases.hashicorp.com/consul/0.7.5/consul_0.7.5_linux_amd64.zip
#unzip consul_0.7.5_linux_amd64.zip
#mv consul /usr/local/bin/consul
#curl -OL https://releases.hashicorp.com/consul/0.7.5/consul_0.7.5_web_ui.zip
#unzip consul_0.7.5_web_ui.zip -d dist
#mv dist /usr/share/consul/ui
