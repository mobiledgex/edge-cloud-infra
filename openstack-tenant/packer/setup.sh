#!/bin/bash
PATH='/usr/bin:/bin'; export PATH

LOGDIR="/etc/mobiledgex"
LOGFILE="${LOGDIR}/creation_log.txt"
DEFAULT_INTERFACE=ens3
ARTIFACTORY_BASEURL='https://artifactory.mobiledgex.net'
MEX_RELEASE=/etc/mex-release

TMPLOG="/var/tmp/creation_log.txt"
exec &> >(tee "$TMPLOG")

[[ "$TRACE" == yes ]] && set -x
set -e

sudo mkdir -p "$LOGDIR"
sudo chmod 700 "$LOGDIR"

# Move log file into 
archive_log() {
	[[ -f "$LOGFILE" ]] \
		&& sudo mv "$LOGFILE" "${LOGFILE}.$( date +'%Y-%m-%d-%H%M' )"
	sudo mv "$TMPLOG" "$LOGFILE"
}
trap 'archive_log' EXIT

if [[ -z "$ROOT_PASS" ]]; then
	echo "Root password not found" >&2
	exit 2
elif [[ -z "$TOTP_KEY" ]]; then
	echo "TOTP key not found" >&2
	exit 2
fi

# Defaults for environment variables
: ${TAG:=master}
: ${ROOT_PASS:=$DEFAULT_ROOT_PASS}
if [[ -z "$INTERFACE" ]]; then
	INTERFACE=$( ls -d /sys/class/net/*/device 2>/dev/null \
			| head -n 1 \
			| cut -d/ -f5 )
	if [[ -z "$INTERFACE" ]]; then
		log "Unable to determine default interface; assuming $DEFAULT_INTERFACE"
		INTERFACE=$DEFAULT_INTERFACE
	fi
fi

log() {
	echo "[$(date)] $*"
}

log_file_contents() {
	[[ -f "$1" ]] || return

	echo "::::::::::  $1  ::::::::::"
	cat "$1"
	echo "::::::::::::::::::::::::::"
	echo
}

die() {
	log "FATAL: $*"
	exit 2
}

download_artifactory_file() {
	local src="$1"
	local dst="$2"
	local mode="$3"
	local arturl="${ARTIFACTORY_BASEURL}/artifactory/baseimage-build/${ARTIFACTORY_ARTIFACTS_TAG}/${src}"
	log "Downloading $arturl"
	sudo curl -s -u "$ARTIFACTORY_CREDS" -o "$dst" "$arturl"
	sudo test -f "$dst" || die "Failed to download file: $arturl"
	[[ -z "$mode" ]] || sudo chmod "$mode" "$dst"
}

# Main
echo "[$(date)] Starting setup.sh for platform \"$OUTPUT_PLATFORM\" ($( pwd ))"

echo "127.0.0.1 $( hostname )" | sudo tee -a /etc/hosts >/dev/null
log_file_contents /etc/hosts

sudo tee /etc/systemd/resolved.conf <<EOT
[Resolve]
DNS=1.1.1.1
FallbackDNS=1.0.0.1
EOT
echo "nameserver 1.1.1.1" | sudo tee -a /etc/resolv.conf >/dev/null
log_file_contents /etc/resolv.conf

log "Downloading files from artifactory for tag $ARTIFACTORY_ARTIFACTS_TAG"
download_artifactory_file ssh.config /root/.ssh/config 600
download_artifactory_file keys/id_rsa_mex /etc/mobiledgex/id_rsa_mex 600
download_artifactory_file keys/id_rsa_mex.pub /tmp/id_rsa_mex.pub
download_artifactory_file keys/id_rsa_mobiledgex.pub /tmp/id_rsa_mobiledgex.pub

log "Setting up SSH"
sudo cp /etc/mobiledgex/id_rsa_mex /root/id_rsa_mex
sudo chmod 600 /root/id_rsa_mex
for SSH_HOME in /root /home/ubuntu; do
	sudo mkdir -p ${SSH_HOME}/.ssh
	sudo cat /tmp/id_rsa_mex.pub /tmp/id_rsa_mobiledgex.pub | sudo tee ${SSH_HOME}/.ssh/authorized_keys
	sudo chmod 700 ${SSH_HOME}/.ssh
	sudo chmod 600 ${SSH_HOME}/.ssh/authorized_keys
	sudo rm -f ${SSH_HOME}/.ssh/known_hosts
done

log "Setting up $MEX_RELEASE"
sudo tee "$MEX_RELEASE" <<EOT
MEX_BUILD="$MEX_BUILD $( TZ=UTC date +'%Y/%m/%d %H:%M %Z' )"
MEX_BUILD_TAG=$TAG
MEX_BUILD_FLAVOR=$FLAVOR
MEX_BUILD_SRC_IMG=$SRC_IMG
MEX_BUILD_SRC_IMG_CHECKSUM=$SRC_IMG_CHECKSUM
MEX_PLATFORM_FLAVOR=$OUTPUT_PLATFORM
EOT

log "Set up docker log file rotation"
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOT
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "50m",
    "max-file": "20"
  }
}
EOT

log "Setting up APT sources"
sudo rm -rf /etc/apt/sources.list.d
sudo tee /etc/apt/sources.list <<EOT
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/packages stratus main
deb https://${APT_USER}:${APT_PASS}@apt.mobiledgex.net stratus-deps main
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu bionic main restricted universe multiverse
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu bionic-updates main restricted universe multiverse
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu-security bionic-security main restricted universe multiverse
EOT

log "Disable cloud config overwrite of APT sources"
sudo tee -a /etc/cloud/cloud.cfg <<EOT
# Preserve /etc/apt/sources.list
apt_preserve_sources_list: true
EOT

log "Set up the APT keys"
curl -s https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/api/gpg/key/public | sudo apt-key add -
curl -s https://${APT_USER}:${APT_PASS}@apt.mobiledgex.net/gpg.key | sudo apt-key add -
sudo apt-get update

log "Disable snap"
sudo apt purge -y snapd
sudo rm -rf /snap /var/snap /var/cache/snapd /var/lib/snapd

log "Switch networking back to ifupdown"
sudo apt-get install -y ifupdown
sudo apt-get purge -y netplan.io
echo "source /etc/network/interfaces.d/*.cfg" | sudo tee -a /etc/network/interfaces

log "Install mobiledgex ${TAG#v}"
# avoid interactive for iptables-persistent
echo iptables-persistent iptables-persistent/autosave_v6 boolean true | sudo debconf-set-selections
echo iptables-persistent iptables-persistent/autosave_v4 boolean true | sudo debconf-set-selections
sudo apt-get install -y mobiledgex=${TAG#v}
[[ $? -ne 0 ]] && die "Failed to install extra packages"

if [[ "$OUTPUT_PLATFORM" == vsphere ]]; then
	log "Adding VMWare cloud-init Guestinfo"
	sudo curl  -sSL https://raw.githubusercontent.com/vmware/cloud-init-vmware-guestinfo/v1.3.1/install.sh |sudo sh -
fi

log "dhclient $INTERFACE"
sudo dhclient "$INTERFACE"
ip addr
ip route

log "Enabling the mobiledgex service"
sudo systemctl enable mobiledgex

log "Updating dhclient timeout"
sudo perl -i -p -e s/'timeout 300;'/'timeout 15;'/g /etc/dhcp/dhclient.conf

if [[ "$OUTPUT_PLATFORM" == vsphere ]]; then
	log "Removing serial console from grub"
	sudo perl -i -p -e s/'"console=tty1 console=ttyS0"'/'""'/g /etc/default/grub.d/50-cloudimg-settings.cfg
	sudo grub-mkconfig -o /boot/grub/grub.cfg
fi

log "Setting the root password"
echo "root:$ROOT_PASS" | sudo chpasswd

log "Setting up root TOTP"
sudo apt-get install -y libpam-google-authenticator
echo "auth required pam_google_authenticator.so" \
	| sudo tee -a /etc/pam.d/login
sudo tee /root/.google_authenticator >/dev/null <<EOT
$TOTP_KEY
" RATE_LIMIT 3 30
" WINDOW_SIZE 17
" DISALLOW_REUSE
" TOTP_AUTH
EOT
sudo chmod 400 /root/.google_authenticator

log "System setup"
sudo swapoff -a
sudo sed -i "s/cgroup-driver=systemd/cgroup-driver=cgroupfs/g" /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo kubeadm config images pull
sudo usermod -aG docker root

echo "d /run/sshd 0755 root root" | sudo tee -a /usr/lib/tmpfiles.d/sshd.conf
sudo rm -f /etc/systemd/system/ssh.service
sudo tee /etc/systemd/system/ssh.service <<'EOT'
[Unit]
Description=OpenBSD Secure Shell server
After=network.target auditd.service
ConditionPathExists=!/etc/ssh/sshd_not_to_be_run

[Service]
EnvironmentFile=-/etc/default/ssh
ExecStartPre=/usr/sbin/sshd -t
ExecStart=/usr/sbin/sshd -D $SSHD_OPTS
ExecReload=/usr/sbin/sshd -t
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
Type=notify

[Install]
WantedBy=multi-user.target
Alias=sshd.service
EOT

sudo tee /etc/cron.hourly/sshd-stale-session-cleanup <<'EOT'
#!/bin/sh
systemctl 2>/dev/null \
    | grep 'scope.*abandoned' \
    | awk '{print $1}' \
    | sudo xargs -r systemctl stop >>/var/tmp/session-cleanup.log 2>&1
EOT
sudo chmod +x /etc/cron.hourly/sshd-stale-session-cleanup

# Create Chef related files required during cloud-init
sudo mkdir -p /etc/chef
sudo touch /etc/chef/client.rb

# systemd unit file for chef-client
sudo tee /etc/default/chef-client <<'EOT'
# Chef client config file
CONFIG=/etc/chef/client.rb

# Interval in seconds
INTERVAL=600

# Splay interval in seconds
SPLAY=20

# Other options
OPTIONS="-d 1 --chef-license accept"
EOT

sudo tee /etc/systemd/system/chef-client.service <<'EOT'
[Unit]
Description = Chef Client daemon
After = network.target auditd.service

[Service]
Type = forking
EnvironmentFile = /etc/default/chef-client
PIDFile = /var/run/chef/client.pid
ExecStart = /usr/bin/chef-client -c $CONFIG -i $INTERVAL -s $SPLAY $OPTIONS
ExecReload = /bin/kill -HUP $MAINPID
SuccessExitStatus = 3
Restart = always

[Install]
WantedBy = multi-user.target
EOT

log "Enabling the chef-client service"
sudo systemctl enable chef-client

if [[ "$OUTPUT_PLATFORM" == vsphere ]]; then
	sudo tee /lib/systemd/system/open-vm-tools.service <<'EOT'
[Unit]
Description=Service for virtual machines hosted on VMware
Documentation=http://open-vm-tools.sourceforge.net/about.php
ConditionVirtualization=vmware
DefaultDependencies=no
Requires=dbus.socket
After=dbus.socket
[Service]
ExecStart=/usr/bin/vmtoolsd
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target
EOT

	log "Enabling the open-vm-tools service"
	sudo systemctl enable open-vm-tools
fi

log "Cleanup"
sudo apt-get autoremove -y

echo "[$(date)] Done setup.sh ($( pwd ))"
