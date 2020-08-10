#!/bin/bash
# this is run at system init time
# TODO: mark so that it does not run again 

set -x

. /etc/mex-release

if [[ "$MEX_PLATFORM_FLAVOR" == vsphere ]]; then
	systemctl status open-vm-tools > /var/log/openvmtool.status.log
	systemctl start open-vm-tools
fi

INIT_COMPLETE_FLAG=/etc/mobiledgex/init-complete
if [[ -f "$INIT_COMPLETE_FLAG" ]]; then
	echo "Already initialized; nothing to do" >&2
	exit 2
fi

umask 027
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
METADIR="$MCONF/openstack/latest"
METADATA="$METADIR/meta_data.json"
NETDATA="$METADIR/network_data.json"
VMWARE_CLOUDINIT=/etc/cloud/cloud.cfg.d/99-DataSourceVMwareGuestInfo.cfg

# Main
log "Starting mobiledgex init"

if [[ -f "$VMWARE_CLOUDINIT" ]]; then
        log "VMware cloud-init case, fetch metadata from vmtoolsd"
        # check that metadata exists, if it does not then exit.
        if ! vmtoolsd --cmd "info-get guestinfo.metadata";
        then
            log "VMware metadata is empty, quitting"
            exit 0
        fi
        log "show userdata"
        vmtoolsd --cmd "info-get guestinfo.userdata" > /var/log/userdata.log
        log "VMware cloud-init case, fetch metadata from vmtoolsd"
        mkdir -p $METADIR
        vmtoolsd --cmd "info-get guestinfo.metadata"|base64 -d|python3 -c 'import sys, yaml, json; json.dump(yaml.load(sys.stdin), sys.stdout)' > $METADATA 
fi

mkdir -p $MCONF
mount `blkid -t LABEL="config-2" -odevice` $MCONF

# Load parameters
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
	log "Setting 1.1.1.1 as nameserver"
	echo "nameserver 1.1.1.1" >/etc/resolv.conf
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
