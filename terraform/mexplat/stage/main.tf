provider "azurerm" {
  version = "=1.24.0"

  client_id       = "${var.azure_terraform_service_principal_id}"
  client_secret   = "${var.azure_terraform_service_principal_secret}"
  subscription_id = "${var.azure_subscription_id}"
  tenant_id       = "${var.azure_tenant_id}"

}

provider "google" {
  version = "=2.5.1"

  project = "${var.gcp_project}"
  zone    = "${var.gcp_zone}"
}

provider "cloudflare" {
  version = "=1.14.0"

  email   = "${var.cloudflare_account_email}"
  token   = "${var.cloudflare_account_api_token}"
}

terraform {
  backend "azurerm" {
    storage_account_name  = "mexterraformstate"
    container_name        = "mexplat-tfstate"
    key                   = "stage.tfstate"
  }
}

# Common VM for gitlab, crm, mc, vault, postgres
module "gitlab" {
  source              = "../../modules/vm_gcp"

  instance_name       = "${var.gitlab_instance_name}"
  zone                = "${var.gcp_zone}"
  boot_disk_size      = 100
  tags                = [
    "mexplat-${var.environ_tag}",
    "gitlab-registry",
    "http-server",
    "https-server",
    "pg-5432",
    "crm",
    "mc",
    "stun-turn",
    "vault-ac",
    "${module.fw_vault_gcp.target_tag}"
  ]
  labels              = {
    "environ"         = "${var.environ_tag}",
    "gitlab"          = "true",
    "vault"           = "true",
  }
  ssh_public_key_file = "${var.ssh_public_key_file}"
}

module "vault_b" {
  source              = "../../modules/vm_gcp"

  instance_name       = "${var.vault_b_instance_name}"
  instance_size       = "custom-1-7680-ext"
  zone                = "${var.vault_b_gcp_zone}"
  boot_disk_size      = 20
  tags                = [
    "vault-ac",
    "${module.fw_vault_gcp.target_tag}"
  ]
  labels              = {
    "environ"         = "${var.environ_tag}",
    "vault"           = "true",
  }
  ssh_public_key_file = "${var.ssh_public_key_file}"
}

module "gitlab_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.gitlab_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

module "docker_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.gitlab_docker_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

module "crm_vm_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.crm_vm_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

module "mc_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.mc_vm_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

module "postgres_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.postgres_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

# VM for console
module "console" {
  source              = "../../modules/vm_gcp"

  instance_name       = "${var.console_instance_name}"
  zone                = "${var.gcp_zone}"
  boot_disk_size      = 100
  tags                = [ "http-server", "https-server", "console-debug", "mc" ]
  ssh_public_key_file = "${var.ssh_public_key_file}"
}

module "console_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.console_domain_name}"
  ip                            = "${module.console.external_ip}"
}

module "vault_a_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.vault_a_domain_name}"
  ip                            = "${module.gitlab.external_ip}"
}

module "vault_b_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.vault_b_domain_name}"
  ip                            = "${module.vault_b.external_ip}"
}

module "fw_vault_gcp" {
  source                        = "../../modules/fw_vault_gcp"
  firewall_name                 = "${var.environ_tag}-vault-fw-hc-and-proxy"
  target_tag                    = "${var.environ_tag}-vault-hc-and-proxy"
}
