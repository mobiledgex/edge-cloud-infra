echo starting setup.sh  | sudo tee -a /root/creation_log.txt
whoami  | sudo tee -a /root/creation_log.txt
pwd  | sudo tee -a /root/creation_log.txt
echo 127.0.1.1 `hostname` | sudo tee -a /etc/hosts
cat /etc/hosts  | sudo tee -a /root/creation_log.txt
echo nameserver 1.1.1.1 | sudo tee -a /etc/resolv.conf
cat /etc/resolv.conf  | sudo tee -a /root/creation_log.txt
sudo dhclient ens3  | sudo tee -a /root/creation_log.txt
ip a  | sudo tee -a /root/creation_log.txt
ip r  | sudo tee -a /root/creation_log.txt
sudo apt-get update
sudo apt-get install -y jq
sudo curl -s -o /usr/local/bin/mobiledgex-init.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/mobiledgex-init.sh 
sudo chmod a+rx /usr/local/bin/mobiledgex-init.sh
echo copied mobiledgex-init.sh  | sudo tee -a /root/creation_log.txt
sudo curl -s -o /etc/systemd/system/mobiledgex.service https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/mobiledgex.service
echo copied mobiledgex.serivce  | sudo tee -a /root/creation_log.txt
sudo chmod a+rx /etc/systemd/system/mobiledgex.service
sudo systemctl enable mobiledgex
echo enabled mobiledgex service  | sudo tee -a /root/creation_log.txt
#sudo mkdir -p /root/.ssh
#sudo ls -al /root
#sudo ls -al /root/.ssh
sudo curl -s -o /tmp/id_rsa_mex.pub https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/id_rsa_mex.pub
sudo curl -s -o /tmp/id_rsa_mobiledgex.pub https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/id_rsa_mobiledgex.pub
sudo cat /tmp/id_rsa_mex.pub /tmp/id_rsa_mobiledgex | sudo tee  /root/.ssh/authorized_keys
sudo chmod 700 /root/.ssh
sudo chmod 600 /root/.ssh/authorized_keys
sudo curl -s -o /root/.ssh/config https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/ssh.config
sudo chmod 600 /root/.ssh/config
sudo rm /root/.ssh/known_hosts
echo set up ssh  | sudo tee -a /root/creation_log.txt
#sudo cat /tmp/id_rsa_mex.pub | sudo tee -a ~ubuntu/.ssh/authorized_keys
#sudo ls -alR ~ubuntu/
sudo curl -s -o /root/install-k8s-base.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-base.sh
sudo chmod a+rx /root/install-k8s-base.sh
sudo curl -s -o /root/install-k8s-master.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-master.sh
sudo chmod a+rx /root/install-k8s-master.sh
sudo curl -s -o /root/install-k8s-node.sh https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/install-k8s-node.sh
sudo chmod a+rx /root/install-k8s-node.sh
echo copied k8s install scripts  | sudo tee -a /root/creation_log.txt
#sudo ls -alR /root
#sudo sed -e 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config > /tmp/xxx
#sudo mv /tmp/xxx /etc/ssh/sshd_config
sudo sed -e 's/UsePAM yes/UsePAM no/' /etc/ssh/sshd_config | sudo tee /tmp/sshd_config
mv /tmp/sshd_config /etc/ssh/sshd_config
sudo sed -e 's/ChallengeResponseAuthentication yes/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config | sudo tee /tmp/sshd_config
mv /tmp/sshd_config /etc/ssh/sshd_config
sudo sed -e 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config | sudo tee /tmp/sshd_config
mv /tmp/sshd_config /etc/ssh/sshd_config
sudo chmod 644 /etc/ssh/sshd_config
echo secured sshd_config | sudo tee -a /root/creation_log.txt
echo root:sandhill | sudo chpasswd
echo set root passwd  | sudo tee -a /root/creation_log.txt
#sudo useradd -m -s /bin/bash mobiledgex
#sudo mkdir -p /home/mobiledgex/.ssh
#sudo chown mobiledgex /home/mobiledgex
#sudo chown mobiledgex /home/mobiledgex/.ssh
#sudo usermod -aG sudo mobiledgex
#echo mobiledgex:sandhill | sudo chpasswd
#sudo mkdir -p /home/mobiledgex/.ssh
#sudo cat /tmp/id_rsa_mex.pub | sudo tee -a ~mobiledgex/.ssh/authorized_keys
#sudo chown mobiledgex ~mobiledgex/.ssh/authorized_keys
#echo 'mobiledgex ALL=(ALL:ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers
#sudo cat /etc/ssh/sshd_config
echo starting install of k8s base | sudo tee -a /root/creation_log.txt
sudo sh -x /root/install-k8s-base.sh | sudo tee -a /root/creation_log.txt
sudo chmod a+rw /var/run/docker/sock
echo installed k8s base | sudo tee -a /root/creation_log.txt
#curl -L https://github.com/docker/compose/releases/download/1.22.0/docker-compose-Linux-x86_64 -o /usr/local/bin/docker-compose
sudo curl  https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/docker-compose -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
echo installed docker-compose | sudo tee -a /root/creation_log.txt
#curl -s -o /tmp/helm.tar.gz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz
sudo curl -s -o /tmp/helm.tar.gz https://mobiledgex:sandhill@registry.mobiledgex.net:8000/mobiledgex/helm-v2.11.0.tar.gz
sudo tar xvf /tmp/helm.tar.gz
sudo mv linux-amd64/helm /usr/local/bin/
sudo chmod a+rx /usr/local/bin/helm
echo installed helm | sudo tee -a /root/creation_log.txt
sudo cat /etc/ssh/sshd_config | sudo tee -a  /root/creation_log.txt
echo created at `date` | sudo tee -a /root/creation_log.txt
