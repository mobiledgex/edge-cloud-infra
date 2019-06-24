#!/bin/sh
set -x

sudo apt-get update && apt-get install -y apt-transport-https curl unzip python
sudo apt-get install apt-transport-https ca-certificates curl software-properties-common
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
sudo apt-get update && sudo apt-get install -y docker-ce

sudo which docker
if [ $? -ne 0 ]; then
    echo docker install failed
    exit 1
fi
