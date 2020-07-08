#!/bin/sh

# exit immediately on failure
set -e

# Make sure required binaries exists
type chef-client > /dev/null
[[ $? -ne 0 ]] && die "Missing 'chef-client'. Install chef workstation from https://downloads.chef.io/chef-workstation"
type knife > /dev/null
[[ $? -ne 0 ]] && die "Missing 'knife'. Install chef workstation from https://downloads.chef.io/chef-workstation"
type chef-zero > /dev/null
[[ $? -ne 0 ]] && die "Missing 'chef-zero'. Install it using 'chef gem install chef-zero'"

ROOTPATH=$GOPATH/src/github.com/mobiledgex/edge-cloud-infra
CHEFPATH=$ROOTPATH/chef
CHEFTESTPATH=$ROOTPATH/e2e-tests/chef
KNIFECFG=$CHEFTESTPATH/knife_local.rb
VALIDATIONKEY=/tmp/validation_key.pem

knife opc org create mobiledgex MobiledgeX Org --filename $VALIDATIONKEY -c $KNIFECFG

# Upload dependent cookbooks
for remoteCookbook in "docker" "iptables"; do
  if [ ! -f /tmp/chef_$remoteCookbook.tar.gz ]; then
    knife supermarket download $remoteCookbook -f /tmp/chef_$remoteCookbook.tar.gz
  fi
  tar -xzf /tmp/chef_$remoteCookbook.tar.gz -C /tmp/
  knife cookbook upload $remoteCookbook -c $KNIFECFG --cookbook-path /tmp/
  rm -r /tmp/$remoteCookbook
done

# Upload cookbooks from our repository
for cookbook in `ls $CHEFPATH/cookbooks/`
do
  echo "Upload cookbook $cookbook"
  if [ -d "$CHEFPATH/cookbooks/$cookbook" ]; then
    echo "Upload cookbook $cookbook"
    knife cookbook upload $cookbook -c $KNIFECFG
  fi
done

policyGroup="local"
for policyFile in `ls $CHEFPATH/policyfiles/*.lock.json`
do
  echo "Upload policy $policyFile to group $policyGroup"
  ( cd $CHEFPATH/policyfiles ; chef push $policyGroup $policyFile -c $KNIFECFG )
done
