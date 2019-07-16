#!/bin/sh

# exit immediately on failure (enable approle will fail if already enabled)
set -e

# Set up the profiles for the edge-cloud approles.
# This assumes a global Vault for all regions, so paths in the Vault
# are region-specific.
# This script should be run for each new region that we bring online.

# You may need to set the following env vars before running:
# VAULT_ADDR=http://127.0.0.1:8200
# VAULT_TOKEN=<my auth token>

# Region should be set to the correct region name
# REGION=local
REGION=$1

if [ -z "$REGION" ]; then
    echo "Usage: setup-region.sh <region>"
    exit 1
fi
echo "Setting up Vault region $REGION"

# enable approle auth if not already enabled
auths=$(vault auth list)
case "$auths" in
    *_"approle"_*) ;;
    *) vault auth enable approle
esac

# shepherd approle
# This has access to all influxdb accounts
cat >/tmp/shepherd-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "secret/data/$REGION/accounts/influxdb" {
  capabilities = [ "read" ]
}
EOF
vault policy write $REGION.shepherd /tmp/shepherd-pol.hcl
rm /tmp/shepherd-pol.hcl
vault write auth/approle/role/$REGION.shepherd period="720h" policies="$REGION.shepherd"
# get shepherd app roleID and generate secretID
vault read auth/approle/role/$REGION.shepherd/role-id
vault write -f auth/approle/role/$REGION.shepherd/secret-id

# generate secret string:
# openssl rand -base64 128
