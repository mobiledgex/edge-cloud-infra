#!/bin/bash
PATH='/usr/bin:/bin:/usr/sbin:/sbin'; export PATH

log() {
	echo
	echo "================================================================================"
	echo "$*"
}

log "remove old kernels"
echo grub grub/update_grub_changeprompt_threeway select keep_current | sudo debconf-set-selections
echo grub-legacy-ec2 grub/update_grub_changeprompt_threeway select keep_current | sudo debconf-set-selections
NEED_KERNEL=$( apt depends linux-image-virtual 2>/dev/null \
	| grep 'Depends: linux-image-' | awk '{print $2}' )
for OLD_KERNEL in $( dpkg -l | grep 'linux-image-.*-generic' \
	| awk '{print $2}' | grep -v "^${NEED_KERNEL}$" ); do
	OLD_MODULES=$( echo "$OLD_KERNEL" | sed 's/image/modules/' )
	echo "deleting $OLD_KERNEL and $OLD_MODULES"
	sudo apt-get purge -y "$OLD_KERNEL" "$OLD_MODULES"
done

log "cleanup package cache"
sudo apt-get autoremove -y
sudo rm -f /var/cache/apt/archives/*.deb

log "Enabling the mobiledgex service"
sudo systemctl enable mobiledgex

# IMPORTANT:
# This needs to be the very last thing that runs in the base image build
log "cleanup sudoers file"
sudo rm -f /etc/sudoers.d/90-cloud-init-users 
