provider "azurerm" {
  version = "~> 2.39.0"

  client_id       = var.azure_terraform_service_principal_id
  client_secret   = var.azure_terraform_service_principal_secret
  subscription_id = var.azure_subscription_id
  tenant_id       = var.azure_tenant_id
}

provider "google" {
  version = "=2.5.1"

  project = var.gcp_project
  zone    = var.gcp_zone
}

provider "cloudflare" {
  version = "~> 2.14.0"

  email   = var.cloudflare_account_email
  api_key = var.cloudflare_account_api_token
}

provider "template" {
  version = "~> 2.2"
}

terraform {
  backend "azurerm" {
    storage_account_name = "mexterraformstate"
    container_name       = "mexplat-tfstate"
    key                  = "main.tfstate"
  }
}

module "gitlab" {
  source = "../../modules/vm_gcp"

  instance_name             = var.gitlab_instance_name
  environ_tag               = var.environ_tag
  zone                      = var.gitlab_gcp_zone
  boot_image                = ""
  boot_disk_size            = 100
  allow_stopping_for_update = true
  tags = [
    "mexplat-${var.environ_tag}",
    "gitlab-registry",
    "http-server",
    "https-server",
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
    module.teleport_firewall.target_tag,
  ]
  labels = {
    "environ" = var.environ_tag
    "gitlab"  = "true"
    "owner"   = "ops"
  }
}

# Vault VMs
module "vault_a" {
  source = "../../modules/vm_gcp"

  instance_name  = var.vault_a_vm_name
  environ_tag    = var.environ_tag
  zone           = var.vault_a_gcp_zone
  boot_disk_size = 100
  tags = [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ]
  labels = {
    "environ" = var.environ_tag
    "vault"   = "true"
    "owner"   = "ops"
  }
}

module "vault_a_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.vault_a_domain_name
  ip       = module.vault_a.external_ip
}

module "vault_b" {
  source = "../../modules/vm_gcp"

  instance_name  = var.vault_b_vm_name
  environ_tag    = var.environ_tag
  zone           = var.vault_b_gcp_zone
  boot_disk_size = 100
  tags = [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ]
  labels = {
    "environ" = var.environ_tag
    "vault"   = "true"
    "owner"   = "ops"
  }
}

module "vault_b_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.vault_b_domain_name
  ip       = module.vault_b.external_ip
}

module "vault_c" {
  source              = "../../modules/vm_gcp"

  instance_name       = "${var.vault_c_vm_name}"
  environ_tag         = "${var.environ_tag}"
  zone                = "${var.vault_c_gcp_zone}"
  boot_disk_size      = 100
  tags                = [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ]
  labels              = {
    "environ"         = "${var.environ_tag}",
    "vault"           = "true",
    "owner"           = "ops",
  }
}

module "vault_c_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = "${var.vault_c_domain_name}"
  ip                            = "${module.vault_c.external_ip}"
}

# VM for console
module "console" {
  source = "../../modules/vm_gcp"

  instance_name  = var.console_instance_name
  environ_tag    = var.environ_tag
  instance_size  = "custom-1-7680-ext"
  zone           = var.gcp_zone
  boot_disk_size = 100
  tags = concat([
    "http-server",
    "https-server",
    "console-debug",
    "mc-artifactory",
    "notifyroot",
    "alertmanager",
    "stun-turn",
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
    module.teleport_firewall.target_tag,
  ], module.mc_firewall.target_tags)
  labels = {
    "environ" = var.environ_tag
    "console" = "true"
    "owner"   = "ops"
  }
}

module "console_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.console_domain_name
  ip       = module.console.external_ip
}

module "console_vnc_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.console_vnc_domain_name
  ip       = module.console.external_ip
}

module "alertmanager_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.alertmanager_domain_name
  ip       = module.console.external_ip
}

module "notifyroot_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.notifyroot_domain_name
  ip       = module.console.external_ip
}

module "stun_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.stun_domain_name
  ip       = module.console.external_ip
}

module "fw_vault_gcp" {
  source        = "../../modules/fw_vault_gcp"
  firewall_name = "${var.environ_tag}-vault-fw-hc-and-proxy"
  target_tag    = "${var.environ_tag}-vault-hc-and-proxy"
}

module "mc_firewall" {
  source      = "../../modules/fw_mc_gcp"
  environ_tag = var.environ_tag
  gitlab_ip   = module.gitlab.external_ip
}

module "teleport_firewall" {
  source      = "../../modules/fw_teleport_node_gcp"
  environ_tag = var.environ_tag
}

resource "google_compute_address" "harbor" {
  name    = var.harbor_static_address_name
}

module "harbor" {
  source = "../../modules/vm_gcp"

  instance_name  = var.harbor_instance_name
  environ_tag    = var.environ_tag
  zone           = var.gcp_zone
  boot_disk_size = 50
  tags = [
    "http-server",
    "https-server",
  ]
  labels = {
    "environ" = var.environ_tag
    "console" = "true"
    "owner"   = "ops"
  }
  nat_ip = "${google_compute_address.harbor.address}"
}

module "harbor_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = var.harbor_domain_name
  ip       = module.harbor.external_ip
}
