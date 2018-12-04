# vault data for cloudlets

one directory per cloudlet. for example, hamburg, berlin, bonn, munich. Each dir has openrc.json. The json file is uploaded to vault via update.sh script here.

mexenv.json is also updated to vault.

To retrieve data, set VAULT_ADDR and VAULT_TOKEN as specified in update.sh and run

```
curl  --header "X-Vault-Token: $VAULT_TOKEN"  $VAULT_ADDR/v1/secret/data/...
```
