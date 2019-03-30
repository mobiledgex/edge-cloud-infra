#!/bin/bash
PATH='/usr/bin:/bin'; export PATH

LOGDIR="/etc/mobiledgex"
LOGFILE="${LOGDIR}/creation_log.txt"
DEFAULT_INTERFACE=ens3
ARTIFACTORY_BASEURL='https://artifactory.mobiledgex.net'
DEFAULT_ROOT_PASS=sandhill
DEFAULT_HELM_VERSION=v2.11.0
MEX_RELEASE=/etc/mex-release

TMPLOG="/var/tmp/creation_log.txt"
exec &> >(tee "$TMPLOG")

[[ "$TRACE" == yes ]] && set -x

sudo mkdir -p "$LOGDIR"
sudo chmod 700 "$LOGDIR"

# Move log file into 
archive_log() {
	sudo mv "$LOGFILE" "${LOGFILE}.$( date +'%Y-%m-%d-%H%M' )" 2>/dev/null
	sudo mv "$TMPLOG" "$LOGFILE"
}
trap 'archive_log' EXIT

# Defaults for environment variables
: ${TAG:=master}
: ${ROOT_PASS:=$DEFAULT_ROOT_PASS}
: ${HELM_VERSION:=$DEFAULT_HELM_VERSION}
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
	local arturl="${ARTIFACTORY_BASEURL}/artifactory/baseimage-build/$TAG/$src"
	log "Downloading $arturl"
	sudo curl -s -u "$ARTIFACTORY_CREDS" -o "$dst" "$arturl"
	sudo test -f "$dst" || die "Failed to download file: $arturl"
	[[ -n "$mode" ]] && sudo chmod "$mode" "$dst"
}

# Main
echo "[$(date)] Starting setup.sh ($( pwd ))"

echo "127.0.0.1 $( hostname )" | sudo tee -a /etc/hosts >/dev/null
log_file_contents /etc/hosts

echo "nameserver 1.1.1.1" | sudo tee -a /etc/resolv.conf >/dev/null
log_file_contents /etc/resolv.conf

log "Setting up $MEX_RELEASE"
sudo tee "$MEX_RELEASE" <<EOT
MEX_BUILD="$MEX_BUILD $( TZ=UTC date +'%Y/%m/%d %H:%M %Z' )"
MEX_BUILD_TAG=$TAG
MEX_BUILD_FLAVOR=$FLAVOR
MEX_BUILD_SRC_IMG=$SRC_IMG
MEX_BUILD_SRC_IMG_CHECKSUM=$SRC_IMG_CHECKSUM
EOT

log "Installing extra packages"
sudo apt-get update
sudo apt-get install -y \
	ipvsadm=1:1.28-3 \
	jq=1.5+dfsg-1ubuntu0.1
[[ $? -ne 0 ]] && die "Failed to install extra packages"

log "dhclient $INTERFACE"
sudo dhclient "$INTERFACE"
ip addr
ip route

log "Downloading files from artifactory for tag $TAG"
download_artifactory_file holepunch /etc/mobiledgex/holepunch a+rx
download_artifactory_file holepunch.json /etc/mobiledgex/holepunch.json a+r
download_artifactory_file mobiledgex-init.sh /usr/local/bin/mobiledgex-init.sh a+rx
download_artifactory_file mobiledgex.service /etc/systemd/system/mobiledgex.service a+r
download_artifactory_file docker-compose /usr/local/bin/docker-compose a+rx
download_artifactory_file helm-${HELM_VERSION}.tar.gz /tmp/helm.tar.gz a+r

download_artifactory_file ssh.config /root/.ssh/config 600
download_artifactory_file keys/id_rsa_mex /etc/mobiledgex/id_rsa_mex 600
download_artifactory_file keys/id_rsa_mex.pub /tmp/id_rsa_mex.pub
download_artifactory_file keys/id_rsa_mobiledgex.pub /tmp/id_rsa_mobiledgex.pub

download_artifactory_file install-k8s-base.sh /etc/mobiledgex/install-k8s-base.sh a+rx
download_artifactory_file install-k8s-master.sh /etc/mobiledgex/install-k8s-master.sh a+rx
download_artifactory_file install-k8s-node.sh /etc/mobiledgex/install-k8s-node.sh a+rx

log "Setting up SSH"
sudo cp /etc/mobiledgex/id_rsa_mex /root/id_rsa_mex
sudo chmod 600 /root/id_rsa_mex
sudo mkdir -p /root/.ssh
sudo cat /tmp/id_rsa_mex.pub /tmp/id_rsa_mobiledgex.pub | sudo tee /root/.ssh/authorized_keys
sudo chmod 700 /root/.ssh
sudo chmod 600 /root/.ssh/authorized_keys
sudo rm -f /root/.ssh/known_hosts

log "Enabling the mobiledgex service"
sudo systemctl enable mobiledgex

log "Setting the root password"
echo "root:$ROOT_PASS" | sudo chpasswd

log "Installing k8s base"
sudo env ARTIFACTORY_CREDS="$ARTIFACTORY_CREDS" TAG="$TAG" sh -x /etc/mobiledgex/install-k8s-base.sh 
sudo chmod a+rw /var/run/docker/sock
sudo groupadd docker
sudo usermod -aG docker root

log "Installing helm $HELM_VERSION"
tar xf /tmp/helm.tar.gz linux-amd64/helm
sudo mv linux-amd64/helm /usr/local/bin/helm
sudo chmod a+rx /usr/local/bin/helm

echo "[$(date)] Done setup.sh ($( pwd ))"
