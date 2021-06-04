#!/bin/bash
# must be run as root

[[ "$TRACE" == yes ]] && set -x

USAGE="usage: $( basename $0 ) <options>

 -n <driver-name>		GPU driver name
 -d <driver-path>		GPU driver path
 -l <license-config-path>	GPU driver license config path
 -t <driver-type>		GPU driver type

 -h				Display this help message
"
while getopts ":hn:d:l:t:" OPT; do
        case "$OPT" in
        h) echo "$USAGE"; exit 0 ;;
        n) DRIVERNAME="$OPTARG" ;;
        d) DRIVERPATH="$OPTARG" ;;
        l) LICENSECFGPATH="$OPTARG" ;;
        t) DRIVERTYPE="$OPTARG" ;;
        esac
done
shift $(( OPTIND - 1 ))

die() {
        echo "ERROR: $*" >&2
        exit 2
}

TypePassthrough="GpuTypePassthrough"
TypeVGPU="GpuTypeVgpu"

[[ -z $DRIVERNAME ]] && die "Missing GPU driver name"
[[ -z $DRIVERPATH ]] && die "Missing GPU driver path"
[[ -z $DRIVERTYPE ]] && die "Missing GPU driver type"

if [[ "$DRIVERTYPE" != $TypePassthrough ]] && [[ "$DRIVERTYPE" != $TypeVGPU ]]; then
	die "Invalid GPU driver type, valid types are '$TypePassthrough', '$TypeVGPU'"
fi

if [[ ! -f $DRIVERPATH ]]; then
	die "GPU driver package file '$DRIVERPATH' does not exist"
fi

if [[ ! -z $LICENSECFGPATH ]] && [[ ! -f $LICENSECFGPATH ]]; then
	die "GPU driver license file '$LICENSECFGPATH' does not exist"
fi

echo ">> Installing GPU driver $DRIVERPATH..."
dpkg -i $DRIVERPATH
[[ $? -ne 0 ]] && die "Failed to install $DRIVERPATH package"
echo ""

echo ">> Verify GPU driver $DRIVERPATH is installed..."
nvidia-smi -L
[[ $? -ne 0 ]] && die "Failed to verify if $DRIVERPATH package is installed"
echo ""

if [[ ! -z $LICENSECFGPATH ]]; then
	echo ">> Setup GPU driver license config $LICENSECFGPATH..."
	cp $LICENSECFGPATH /etc/nvidia/gridd.conf
	service nvidia-gridd restart
	[[ $? -ne 0 ]] && die "Failed to restart nvidia-gridd service"
	echo ""

	echo ">> Verify GPU driver license is acquired..."
	cat /var/log/syslog | grep -i "license.*success"
	[[ $? -ne 0 ]] && die "Failed to verify if nvidia-gridd license is configured"
	echo ""
fi

echo ">> Done setting up GPU driver $DRIVERNAME"
