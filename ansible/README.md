## Set up the environment

* Set up the GCP service account
    * [Create a GCP service account for yourself](https://cloud.google.com/docs/authentication/getting-started)
    * Download the service account JSON file and copy it to `~/.gcp-terraform-service-principal.json`
* Pull the latest version of the [mobiledgex/secrets](https://github.com/mobiledgex/secrets/blob/master/README.md) repo
* Source the ansible environment file in your `.bashrc`/`.zshrc` file
  ```bash
  . ~/.mobiledgex/ansible.env
  ```

## Examples

### Create or update the mexplat staging environment

```bash
cd edge-cloud-infra/terraform/mexplat/stage
make plan
make deploy
```

### Update the mexplat staging environment to version "2019-05-06"

```bash
cd edge-cloud-infra/ansible
ansible-playbook -i staging -e edge_cloud_version=2019-05-06 -e @ansible-mex-vault.yml mexplat.yml
```
