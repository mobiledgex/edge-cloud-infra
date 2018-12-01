#!/bin/bash
# this is run at system init time
# TODO: mark so that it does not run again 
# TODO: check for updates from edgeproxy before running

echo starting mobiledgex init >> /tmp/mobiledgex.log
date >> /tmp/mobiledgex.log
MCONF=/mnt/mobiledgex-config
mkdir $MCONF
mount `blkid -t LABEL="config-2" -odevice` $MCONF
skipinit=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.skipinit | sed -e 's/"//'g`
if [ "$skipinit" != "yes" ]; then
	echo mobiledgex initialization >> /tmp/mobiledgex.log
	ipaddress=`cat $MCONF/openstack/latest/network_data.json |jq .networks[0].ip_address | sed -e 's/"//'g` 
	nettype=`cat $MCONF/openstack/latest/network_data.json |jq .networks[0].type | sed -e 's/"//'g` 
	if [ "$nettype" = "ipv4_dhcp" ]; then
		echo using dhcp >> /tmp/mobiledgex.log
		dhclient ens3
	else
		ifconfig ens3  $ipaddress up
		ifconfig ens3 netmask `cat $MCONF/openstack/latest/network_data.json |jq .networks[0].netmask | sed -e 's/"//'g` up
		edgeproxy=`cat $MCONF/openstack/latest/meta_data.json| jq .meta.edgeproxy | sed -e 's/"//'g` 
		ip route add default via $edgeproxy dev ens3
	fi
	privatenet=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.privatenet | sed -e 's/"//'g`
	privaterouter=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.privaterouter | sed -e 's/"//'g`
	if [ "$privatenet" != "" -a "$privaterouter" != "" ]; then
		echo private route entry $privatenet via $privaterouter >> /tmp/mobiledgex.log
		ip route add $privatenet via $privaterouter dev ens3
	fi
	ifconfig -a >>/tmp/mobiledgex.log
	ip route >> /tmp/mobiledgex.log
	hostname `cat $MCONF/openstack/latest/meta_data.json |jq .name | sed -e 's/"//'g`
	echo hostname `hostname` >> /tmp/mobiledgex.log
	role=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.role | sed -e 's/"//'g`
	echo role $role >> /tmp/mobiledgex.log
	grep -v 127.0.1.1 /etc/hosts > /tmp/hosts
	echo 127.0.1.1 `hostname` >> /tmp/hosts
	cp /tmp/hosts /etc/hosts
	dig google.com| grep 'status: NOERROR'
	if [ $? -ne 0 ]; then
	    echo add 1.1.1.1 as nameserver >> /tmp/mobiledgex.log
	    echo nameserver 1.1.1.1 > /etc/resolv.conf
	fi
	skipk8s=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.skipk8s | sed -e 's/"//'g`
	if [ "$role" = "mex-agent-node" ]; then
		echo "initializing mex agent node" >> /tmp/mobiledgex.log
		/root/install-k8s-base.sh >> /tmp/mobiledgex.log
		chmod a+rw /var/run/docker.sock
		curl -s -o /tmp/helm.tar.gz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz
		tar xvf /tmp/helm.tar.gz
		mv linux-arm64/helm /usr/local/bin/
		chmod a+rx /usr/local/bin/helm
	else 
		if [ "$skipk8s" != "yes" ]; then
			echo doing k8s init >> /tmp/mobiledgex.log
			/root/install-k8s-base.sh >> /tmp/mobiledgex.log
			echo k8s-base installed >> /tmp/mobiledgex.log
			masteraddr=`cat $MCONF/openstack/latest/meta_data.json |jq .meta.k8smaster | sed -e 's/"//'g`
			if [ "$role" = "k8s-master" ]; then
				echo k8s-master init >> /tmp/mobiledgex.log
				/root/install-k8s-master.sh ens3 $masteraddr $ipaddress >> /tmp/mobiledgex.log
				echo k8s-master installed >> /tmp/mobiledgex.log
			elif [ "$role" = "k8s-node" ]; then
				echo k8s-node init >> /tmp/mobiledgex.log
				/root/install-k8s-node.sh  ens3 $masteraddr $ipaddress >> /tmp/mobiledgex.log
				echo k8s-node installed >> /tmp/mobiledgex.log
			else
				echo "role is " $role " and not k8s" >> /tmp/mobiledgex.log
			fi
			echo done k8s init for role $role >> /tmp/mobiledgex.log
		else
			echo skipping k8s init for role $role >> /tmp/mobiledgex.log
		fi
	fi
	echo $role > /etc/mobiledgex-role.txt
	echo done mobiledgex init >> /tmp/mobiledgex.log
else
	echo skip mobiledgex init >> /tmp/mobiledgex.log
fi
echo all done exiting >> /tmp/mobiledgex.log
date >> /tmp/mobiledgex.log
