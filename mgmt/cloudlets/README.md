# vault data for cloudlets

VAULT_ADDR is to be set to https://vault.mobiledgex.net
VAULT_TOKEN is to be set to token in vault.json. Usually master root token key at the end of that file.

Having token allows you to get all other secrets.

Notice that this is `cloud-infra` repo.  *Private* repo!
That is why vault.json is here as are other secrets.

one directory per cloudlet. for example, hamburg, berlin, bonn, munich. Each dir has openrc.json. The json file is uploaded to vault via update.sh script here.

mexenv.json is also updated to vault.

To retrieve data, set VAULT_ADDR and VAULT_TOKEN as specified in update.sh and run

```
curl  --header "X-Vault-Token: $VAULT_TOKEN"  $VAULT_ADDR/v1/secret/data/...
```

