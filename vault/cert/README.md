## Vault cert operator for Kubernetes

### Installation

Configure vault parameters (`VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID`) in the manifest.
The vault role needs to have the `certs.read.v1` policy, and preferably nothing apart from that policy.

```
vi vault-cert-operator/vault-cert-operator-deploy.yaml
```

Deploy

```
kubectl create ns vault-cert
kubectl apply -f vault-cert-operator/vault-cert-operator-rbac.yaml
kubectl apply -f vault-cert-operator/crd-vaultcert.yaml
kubectl apply -f vault-cert-operator/vault-cert-operator-deploy.yaml
```

### Obtaining cert for a domain

Create a VaultCert request:

```
kubectl apply -f - <<EOF
apiVersion: stable.mobiledgex.net/v1alpha1
kind: VaultCert
metadata:
  name: test.mobiledgex.net
spec:
  domain:
    - test.mobiledgex.net
  secretName: test-mobiledgex-net-tls
EOF
```

In a few seconds, the operator will create a TLS secret with the name specified
in the VaultCert request:

```
kubectl get secret test-mobiledgex-net-tls
```

NOTE: The cert will be created in the namespace the VaultCert request was created in.

### Troubleshooting

Check the logs of the operator:

```
kubectl -n vault-cert logs -f -l operator=vault-cert-operator
```

## Vault cert sidecar for Docker

### Installation

This will normally be deployed alongside an envoy which has `/etc/envoy/certs`
mounted and an `sds.yaml` similar to one the root LBs use currently.

```
read -p "VAULT_ADDR: " VAULT_ADDR
read -p "VAULT_ROLE_ID: " VAULT_ROLE_ID
read -p "VAULT_SECRET_ID: " VAULT_SECRET_ID
export VAULT_ADDR VAULT_ROLE_ID VAULT_SECRET_ID

mkdir -p ~/tmp/certs
docker run --rm -e VAULT_ADDR -e VAULT_ROLE_ID -e VAULT_SECRET_ID \
	-v certsvol:/etc/envoy/certs \
	-e CERT_DOMAIN=test.mobiledgex.net \
	harbor.mobiledgex.net/mobiledgex/vault-cert-sidecar:latest
```
