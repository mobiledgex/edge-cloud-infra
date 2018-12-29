sudo mkdir -p /etc/mobiledgex
sudo chmod 700 /etc/mobiledgex
echo starting setup.sh  | sudo tee -a /etc/mobiledgex/creation_log.txt
pwd  | sudo tee -a /etc/mobiledgex/creation_log.txt
echo 127.0.1.1 `hostname` | sudo tee -a /etc/hosts
cat /etc/hosts  | sudo tee -a /etc/mobiledgex/creation_log.txt
echo nameserver 1.1.1.1 | sudo tee -a /etc/resolv.conf
cat /etc/resolv.conf  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo dhclient ens3  | sudo tee -a /etc/mobiledgex/creation_log.txt
ip a  | sudo tee -a /etc/mobiledgex/creation_log.txt
ip r  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo apt-get update
sudo apt-get install -y jq ipvsadm
sudo curl -s -o /etc/mobiledgex/holepunch https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/holepunch
sudo curl -s -o /etc/mobiledgex/holepunch.json https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/holepunch.json
sudo chmod a+rx /etc/mobiledgex/holepunch
sudo chmod a+r /etc/mobiledgex/holepunch.json
sudo curl -s -o /usr/local/bin/mobiledgex-init.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/mobiledgex-init.sh 
sudo chmod a+rx /usr/local/bin/mobiledgex-init.sh
echo copied mobiledgex-init.sh  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo curl -s -o /etc/systemd/system/mobiledgex.service https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/mobiledgex.service
echo copied mobiledgex.serivce  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo chmod a+rx /etc/systemd/system/mobiledgex.service
sudo systemctl enable mobiledgex
echo enabled mobiledgex service  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo curl -s -o /tmp/id_rsa_mex.pub https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/id_rsa_mex.pub
sudo curl -s -o /etc/mobiledgex/id_rsa_mex https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/id_rsa_mex
sudo chmod 600 /etc/mobiledgex/id_rsa_mex
sudo cp /etc/mobiledgex/id_rsa_mex /root/id_rsa_mex
sudo chmod 600 /root/id_rsa_mex
sudo curl -s -o /tmp/id_rsa_mobiledgex.pub https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/id_rsa_mobiledgex.pub
sudo cat /tmp/id_rsa_mex.pub /tmp/id_rsa_mobiledgex.pub | sudo tee  /root/.ssh/authorized_keys
sudo chmod 700 /root/.ssh
sudo chmod 600 /root/.ssh/authorized_keys
sudo curl -s -o /root/.ssh/config https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/ssh.config
sudo chmod 600 /root/.ssh/config
sudo rm /root/.ssh/known_hosts
echo set up ssh  | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo curl -s -o /etc/mobiledgex/install-k8s-base.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-base.sh
sudo chmod a+rx /etc/mobiledgex/install-k8s-base.sh
sudo curl -s -o /etc/mobiledgex/install-k8s-master.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-master.sh
sudo chmod a+rx /etc/mobiledgex/install-k8s-master.sh
sudo curl -s -o /etc/mobiledgex/install-k8s-node.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-node.sh
sudo chmod a+rx /etc/mobiledgex/install-k8s-node.sh
echo copied k8s install scripts  | sudo tee -a /etc/mobiledgex/creation_log.txt
echo root:sandhill | sudo chpasswd
echo set root passwd  | sudo tee -a /etc/mobiledgex/creation_log.txt
echo starting install of k8s base | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo sh -x /etc/mobiledgex/install-k8s-base.sh | sudo tee -a /etc/mobiledgex/creation_log.txt
sudo chmod a+rw /var/run/docker/sock
sudo groupadd docker
sudo usermod -aG docker root
echo installed k8s base | sudo tee -a /etc/mobiledgex/creation_log.txt
#curl -L https://github.com/docker/compose/releases/download/1.22.0/docker-compose-Linux-x86_64 -o /usr/local/bin/docker-compose
sudo curl  https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/docker-compose -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
echo installed docker-compose | sudo tee -a /etc/mobiledgex/creation_log.txt
#curl -s -o /tmp/helm.tar.gz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz
sudo curl -s -o /tmp/helm.tar.gz https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/helm-v2.11.0.tar.gz
sudo tar xvf /tmp/helm.tar.gz
sudo mv linux-amd64/helm /usr/local/bin/
sudo chmod a+rx /usr/local/bin/helm
echo installed helm | sudo tee -a /etc/mobiledgex/creation_log.txt
echo created at `date` | sudo tee -a /etc/mobiledgex/creation_log.txt
