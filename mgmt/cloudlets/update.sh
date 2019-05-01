#!/bin/bash

export VAULT_ADDR=https://vault.mobiledgex.net
export VAULT_TOKEN=0eb94cce-ab3f-33a3-bd66-0046e8743a14

for d in hamburg bonn berlin munich frankfurt mpk-tip wwt-atc; do
    curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @$d/openrc.json  $VAULT_ADDR/v1/secret/data/cloudlet/openstack/$d/openrc.json
done

curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @gcp/auth_key.json  $VAULT_ADDR/v1/secret/data/cloudlet/gcp/auth_key.json

curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @mexenv.json  $VAULT_ADDR/v1/secret/data/cloudlet/openstack/mexenv.json
