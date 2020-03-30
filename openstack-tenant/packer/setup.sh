#!/bin/bash
PATH='/usr/bin:/bin'; export PATH

LOGDIR="/etc/mobiledgex"
LOGFILE="${LOGDIR}/creation_log.txt"
DEFAULT_INTERFACE=ens3
ARTIFACTORY_BASEURL='https://artifactory.mobiledgex.net'
DEFAULT_ROOT_PASS=sandhill
MEX_RELEASE=/etc/mex-release

TMPLOG="/var/tmp/creation_log.txt"
exec &> >(tee "$TMPLOG")

[[ "$TRACE" == yes ]] && set -x
set -ex

sudo mkdir -p "$LOGDIR"
sudo chmod 700 "$LOGDIR"

# Move log file into 
archive_log() {
	[[ -f "$LOGFILE" ]] \
		&& sudo mv "$LOGFILE" "${LOGFILE}.$( date +'%Y-%m-%d-%H%M' )"
	sudo mv "$TMPLOG" "$LOGFILE"
}
trap 'archive_log' EXIT

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
echo "[$(date)] Starting setup.sh ($( pwd ))"

echo "127.0.0.1 $( hostname )" | sudo tee -a /etc/hosts >/dev/null
log_file_contents /etc/hosts

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
sudo mkdir -p /root/.ssh
sudo cat /tmp/id_rsa_mex.pub /tmp/id_rsa_mobiledgex.pub | sudo tee /root/.ssh/authorized_keys
sudo chmod 700 /root/.ssh
sudo chmod 600 /root/.ssh/authorized_keys
sudo rm -f /root/.ssh/known_hosts

log "Setting up $MEX_RELEASE"
sudo tee "$MEX_RELEASE" <<EOT
MEX_BUILD="$MEX_BUILD $( TZ=UTC date +'%Y/%m/%d %H:%M %Z' )"
MEX_BUILD_TAG=$TAG
MEX_BUILD_FLAVOR=$FLAVOR
MEX_BUILD_SRC_IMG=$SRC_IMG
MEX_BUILD_SRC_IMG_CHECKSUM=$SRC_IMG_CHECKSUM
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
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu xenial main restricted universe multiverse
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu xenial-updates main restricted universe multiverse
deb https://${APT_USER}:${APT_PASS}@artifactory.mobiledgex.net/artifactory/ubuntu-security xenial-security main restricted universe multiverse
deb https://${APT_USER}:${APT_PASS}@apt.mobiledgex.net/nvidia main main
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

log "Install mobiledgex ${TAG#v}"
# avoid interactive for iptables-persistent
echo iptables-persistent iptables-persistent/autosave_v6 boolean true | sudo debconf-set-selections
echo iptables-persistent iptables-persistent/autosave_v4 boolean true | sudo debconf-set-selections
sudo apt-get install -y mobiledgex=${TAG#v}
[[ $? -ne 0 ]] && die "Failed to install extra packages"

log "dhclient $INTERFACE"
sudo dhclient "$INTERFACE"
ip addr
ip route

log "Enabling the mobiledgex service"
sudo systemctl enable mobiledgex

log "Setting the root password"
echo "root:$ROOT_PASS" | sudo chpasswd

log "System setup"
sudo swapoff -a
sudo sed -i "s/cgroup-driver=systemd/cgroup-driver=cgroupfs/g" /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo kubeadm config images pull
sudo usermod -aG docker root

echo "[$(date)] Done setup.sh ($( pwd ))"
