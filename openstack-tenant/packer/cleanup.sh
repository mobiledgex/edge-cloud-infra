#!/bin/bash
PATH='/usr/bin:/bin:/usr/sbin:/sbin'; export PATH

log() {
	echo
	echo "================================================================================"
	echo "$*"
}

log "cleanup sudoers file"
sudo rm -f /etc/sudoers.d/90-cloud-init-users 
