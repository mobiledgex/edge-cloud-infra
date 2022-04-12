#!/bin/bash
# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


export VAULT_ADDR=https://vault.mobiledgex.net
export VAULT_TOKEN=0eb94cce-ab3f-33a3-bd66-0046e8743a14

for d in hamburg bonn berlin munich frankfurt mpk-tip wwt-atc; do
    curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @$d/openrc.json  $VAULT_ADDR/v1/secret/data/cloudlet/openstack/$d/openrc.json
done

curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @gcp/auth_key.json  $VAULT_ADDR/v1/secret/data/cloudlet/gcp/auth_key.json

curl  --header "X-Vault-Token: $VAULT_TOKEN" --request POST --data @mexenv.json  $VAULT_ADDR/v1/secret/data/cloudlet/openstack/mexenv.json
