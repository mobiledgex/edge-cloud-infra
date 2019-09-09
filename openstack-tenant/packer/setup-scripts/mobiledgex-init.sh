#!/bin/bash
# this is run at system init time
# TODO: mark so that it does not run again 

set -x

INIT_COMPLETE_FLAG=/etc/mobiledgex/init-complete
if [[ -f "$INIT_COMPLETE_FLAG" ]]; then
	echo "Already initialized; nothing to do" >&2
	exit 2
fi

LOGFILE=/var/log/mobiledgex.log
log() {
	if [[ $# -gt 0 ]]; then
		echo "[$(date)] $*" | tee -a "$LOGFILE"
	else
		echo "[$(date)]"
		tee -a "$LOGFILE"
	fi
}

MCONF=/mnt/mobiledgex-config

# Main

log "Starting mobiledgex init"

mkdir -p $MCONF
mount `blkid -t LABEL="config-2" -odevice` $MCONF

# Load parameters
METADATA="$MCONF/openstack/latest/meta_data.json"
NETDATA="$MCONF/openstack/latest/network_data.json"

set_param() {
	local file="$1"
	local var="$2"
	local param="$3"
	local val=$( jq -r "$param // empty" "$file" )
	eval "$var='$val'"
}

# Set variable based on JSON path in metadata file
set_metadata_param() {
	set_param "$METADATA" "$@"
}

# Set variable based on JSON path in network data file
set_network_param() {
	set_param "$NETDATA" "$@"
}

set_metadata_param HOSTNAME .name
set_metadata_param HOLEPUNCH .meta.holepunch
set_metadata_param UPDATE .meta.update
set_metadata_param SKIPINIT .meta.skipinit
set_metadata_param INTERFACE .meta.interface
set_metadata_param ROLE .meta.role
set_metadata_param SKIPK8S .meta.skipk8s
set_metadata_param MASTERADDR .meta.k8smaster
set_metadata_param UPDATEHOSTNAME .meta.updatehostname

set_network_param IPADDR '.networks[0].ip_address'
set_network_param NETMASK '.networks[0].netmask'
set_network_param NETTYPE '.networks[0].type'

if [[ -z "$INTERFACE" ]]; then
	INTERFACE=$( ls -d /sys/class/net/*/device 2>/dev/null \
			| head -n 1 \
			| cut -d/ -f5 )
	[[ -z "$INTERFACE" ]] && INTERFACE=ens3
fi

echo 127.0.0.1 `hostname` >> /etc/hosts
[[ "$UPDATEHOSTNAME" == yes ]] && sed -i "s|^\(127\.0\.1\.1 \).*|\1${HOSTNAME}|" /etc/hosts

if [[ -n "$HOLEPUNCH" ]]; then
	sed -i "s/22222/${HOLEPUNCH}/" /etc/mobiledgex/holepunch.json
	cd /etc/mobiledgex
	/etc/mobiledgex/holepunch write-systemd-file
	systemctl enable holepunch
	systemctl start holepunch
	systemctl status holepunch
fi

usermod -aG docker ubuntu
chmod a+rw /var/run/docker.sock

if [[ "$SKIPINIT" == yes ]]; then
	log "Skipping mobiledgex init as instructed"
	exit 0
fi

ifconfig -a | log
ip route | log

if [[ -z "$ROLE" ]]; then
	log "WARNING: Role is empty"
else
	log "ROLE: $ROLE"
fi

if ! dig google.com | grep 'status: NOERROR' >/dev/null; then
	log "Adding 1.1.1.1 as nameserver"
	echo "nameserver 1.1.1.1" >>/etc/resolv.conf
fi

# TODO: Updates; and also if supported, disable run-once flag check at the top

if [[ "$ROLE" == mex-agent-node ]]; then
	log "Initializing mex agent node"
	systemctl disable kubelet
	systemctl stop kubelet
elif [[ "$SKIPK8S" == yes ]]; then
	log "Skipping k8s init for role $ROLE"
	systemctl disable kubelet
	systemctl stop kubelet
else
	log "K8s init for role $ROLE"
	case "$ROLE" in
	k8s-master)
		sh -x /etc/mobiledgex/install-k8s-master.sh "$INTERFACE" "$MASTERADDR" "$IPADDR" | log
		if [[ "${PIPESTATUS[0]}" != 0 ]]; then
			log "K8s master init failed"
			exit 2
		fi
		;;
	k8s-node)
		sh -x /etc/mobiledgex/install-k8s-node.sh "$INTERFACE" "$MASTERADDR" "$IPADDR" | log
		if [[ "${PIPESTATUS[0]}" != 0 ]]; then
			log "K8s node init failed"
			exit 2
		fi
		;;
	*)
		log "Neither k8s master nor k8s node: $ROLE"
		;;
	esac
	log "Finished k8s init for role $ROLE"
fi

touch "$INIT_COMPLETE_FLAG"
log "Finished mobiledgex init"
