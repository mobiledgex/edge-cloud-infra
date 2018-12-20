#!/bin/bash
set -x
dir=~/src/github.com/mobiledgex/edge-cloud-infra/k8s-prov
dest=registry.mobiledgex.net:files-repo/mobiledgex
for f in   install-k8s-*.sh; do
    scp $dir/$f $dest
done
scp ~/src/github.com/mobiledgex/edge-cloud-infra/openstack-tenant/qcow2/mobiledgex-16.04-qcow2/mobiledgex-init.sh $dest
dir=~/src/github.com/mobiledgex/edge-cloud-infra/mexctl
for d in application cluster platform kustomize; do
    scp -r $dir/$d $dest
done

