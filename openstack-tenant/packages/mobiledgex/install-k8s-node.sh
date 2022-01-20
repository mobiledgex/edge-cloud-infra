#!/bin/sh
# must be run as root
#  on all nodes
set -x
if [ $# -lt 1 ]; then
	echo "Insufficient arguments"
	echo "Need master-ip"
	exit 1
fi
MASTERIP=$1
echo "Master IP $MASTERIP"

systemctl is-active --quiet kubelet
if [ $? -ne 0 ]; then
  systemctl start kubelet
  systemctl enable kubelet
fi

echo installing k8s node, wait...
cd /tmp

curl -sf ${MASTERIP}:20800/k8s-join-cmd >k8s-join-cmd
if [ $? -ne 0 -o ! -s k8s-join-cmd ]; then
	sleep 60
	echo waiting for join-cmd
	curl -sf ${MASTERIP}:20800/k8s-join-cmd >k8s-join-cmd
	while [ $? -ne 0 -o ! -s k8s-join-cmd ]; do
		sleep 7
		curl -sf ${MASTERIP}:20800/k8s-join-cmd >k8s-join-cmd
	done
fi
echo got join cmd
JOIN=`cat /tmp/k8s-join-cmd`
cat k8s-join-cmd
echo running $JOIN --ignore-preflight-errors=all
$JOIN --ignore-preflight-errors=all
echo finished running join
