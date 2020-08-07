#!/bin/bash
PATH='/usr/bin:/bin'; export PATH

log() {
	echo
	echo "================================================================================"
	echo "$*"
}

# Abort on errors
set -e

log "1.1.1.1 Ensure mounting of cramfs filesystems is disabled"
echo "install cramfs /bin/true" | sudo tee /etc/modprobe.d/cramfs.conf

log "1.1.1.2 Ensure mounting of freevxfs filesystems is disabled"
echo "install freevxfs /bin/true" | sudo tee /etc/modprobe.d/freevxfs.conf

log "1.1.1.3 Ensure mounting of jffs2 filesystems is disabled"
echo "install jffs2 /bin/true" | sudo tee /etc/modprobe.d/jffs2.conf

log "1.1.1.4 Ensure mounting of hfs filesystems is disabled"
echo "install hfs /bin/true" | sudo tee /etc/modprobe.d/hfs.conf

log "1.1.1.5 Ensure mounting of hfsplus filesystems is disabled"
echo "install hfsplus /bin/true" | sudo tee /etc/modprobe.d/hfsplus.conf

log "1.1.1.6 Ensure mounting of udf filesystems is disabled"
echo "install udf /bin/true" | sudo tee /etc/modprobe.d/udf.conf

log "1.1.16 Ensure noexec option set on /dev/shm partition"
echo "tmpfs /dev/shm tmpfs defaults,nodev,nosuid,noexec 0 0" \
	| sudo tee -a /etc/fstab

log "1.3.1 Ensure AIDE is installed"
echo "postfix postfix/mailname string localhost" \
	| sudo debconf-set-selections
echo "postfix postfix/main_mailer_type string 'Local only'" \
	| sudo debconf-set-selections
sudo apt-get install -y aide aide-common

log "1.3.2 Ensure filesystem integrity is regularly checked"
sudo rm -f /etc/cron.daily/aide
sudo tee /etc/cron.weekly/aide-check <<'EOT'
#!/bin/sh
/usr/bin/aide.wrapper --config /etc/aide/aide.conf --check >/var/log/aide-check.log 2>&1
EOT

log "1.4.1 Ensure permissions on bootloader config are configured"
sudo chown root:root /boot/grub/grub.cfg
sudo chmod og-rwx /boot/grub/grub.cfg

log "1.5.1 Ensure core dumps are restricted"
echo "* hard core 0" | sudo tee -a /etc/security/limits.conf
echo "fs.suid_dumpable = 0" | sudo tee -a /etc/sysctl.conf

log "1.5.3 Ensure address space layout randomization (ASLR) is enabled"
echo "kernel.randomize_va_space = 2" | sudo tee -a /etc/sysctl.conf


# Final step
log "Initialize the AIDE database"
sudo aideinit
