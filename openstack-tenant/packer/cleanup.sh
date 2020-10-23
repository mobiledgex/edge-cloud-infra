#!/bin/bash
PATH='/usr/bin:/bin:/usr/sbin:/sbin'; export PATH

log() {
	echo
	echo "================================================================================"
	echo "$*"
}

log "cleanup package cache"
sudo apt-get autoremove -y
sudo rm -f /var/cache/apt/archives/*.deb

# IMPORTANT:
# This needs to be the very last thing that runs in the base image build
log "cleanup sudoers file"
sudo rm -f /etc/sudoers.d/90-cloud-init-users 
