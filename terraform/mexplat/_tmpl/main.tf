provider "azurerm" {
  client_id       = var.azure_terraform_service_principal_id
  client_secret   = var.azure_terraform_service_principal_secret
  subscription_id = var.azure_subscription_id
  tenant_id       = var.azure_tenant_id
}

provider "google" {
  project = var.gcp_project
  zone    = var.gcp_zone
}

provider "cloudflare" {
  email   = var.cloudflare_account_email
  api_key = var.cloudflare_account_api_token
}

module "gitlab" {
  source = "../../modules/vm_gcp"

  instance_name             = local.gitlab_instance_name
  environ_tag               = var.environ_tag
  zone                      = var.gitlab_gcp_zone
  boot_image                = ""
  boot_disk_size            = 100
  allow_stopping_for_update = true
  tags = concat(var.global_tags, [
    "mexplat-${var.environ_tag}",
    "gitlab-registry",
    "http-server",
    "https-server",
    module.teleport_firewall.target_tag,
  ])
  labels = {
    "environ" = var.environ_tag
    "gitlab"  = "true"
    "owner"   = "ops"
    "groups"  = "gitlab"
  }
}

# Vault VMs
module "vault_a" {
  source = "../../modules/vm_gcp"

  instance_name  = local.vault_a_vm_name
  environ_tag    = var.environ_tag
  zone           = var.vault_a_gcp_zone
  boot_disk_size = 100
  tags = concat(var.global_tags, [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ])
  labels = {
    "environ" = var.environ_tag
    "vault"   = "true"
    "owner"   = "ops"
    "groups"  = "vault"
  }
}

module "vault_a_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.vault_a_domain_name
  ip       = module.vault_a.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "vault_b" {
  source = "../../modules/vm_gcp"

  instance_name  = local.vault_b_vm_name
  environ_tag    = var.environ_tag
  zone           = var.vault_b_gcp_zone
  boot_disk_size = 100
  tags = concat(var.global_tags, [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ])
  labels = {
    "environ" = var.environ_tag
    "vault"   = "true"
    "owner"   = "ops"
    "groups"  = "vault"
  }
}

module "vault_b_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.vault_b_domain_name
  ip       = module.vault_b.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "vault_c" {
  source              = "../../modules/vm_gcp"

  instance_name       = local.vault_c_vm_name
  environ_tag         = var.environ_tag
  zone                = var.vault_c_gcp_zone
  boot_disk_size      = 100
  tags                = concat(var.global_tags, [
    "mexplat-${var.environ_tag}",
    "vault-ac",
    module.fw_vault_gcp.target_tag,
    module.teleport_firewall.target_tag,
  ])
  labels              = {
    "environ"         = var.environ_tag,
    "vault"           = "true",
    "owner"           = "ops",
    "groups"          = "vault"
  }
}

module "vault_c_dns" {
  source                        = "../../modules/cloudflare_record"
  hostname                      = local.vault_c_domain_name
  ip                            = module.vault_c.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

# VM for console
module "console" {
  source = "../../modules/vm_gcp"

  instance_name  = local.console_instance_name
  environ_tag    = var.environ_tag
  instance_size  = "custom-1-7680-ext"
  zone           = var.gcp_zone
  boot_disk_size = 100
  tags = concat(var.global_tags, [
    "alertmanager",
    "console-debug",
    "http-server",
    "https-server",
    "jaeger",
    "mc-artifactory",
    "notifyroot",
    "restricted-ssh",
    "restricted-ssh-overrides",
    "stun-turn",
    module.teleport_firewall.target_tag,
  ], module.mc_firewall.target_tags)
  labels = {
    "environ" = var.environ_tag
    "console" = "true"
    "owner"   = "ops"
    "groups"  = "alertmanager,console,esproxy,jaeger,notifyroot"
  }
}

module "console_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.console_domain_name
  ip       = module.console.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "console_vnc_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.console_vnc_domain_name
  ip       = module.console.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "alertmanager_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.alertmanager_domain_name
  ip       = module.console.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "notifyroot_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.notifyroot_domain_name
  ip       = module.console.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "stun_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.stun_domain_name
  ip       = module.console.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
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
  name    = local.harbor_static_address_name
}

module "harbor" {
  source = "../../modules/vm_gcp"

  instance_name  = local.harbor_instance_name
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
  nat_ip = google_compute_address.harbor.address
}

module "harbor_dns" {
  source   = "../../modules/cloudflare_record"
  hostname = local.harbor_domain_name
  ip       = module.harbor.external_ip
  cloudflare_zone_id = var.cloudflare_zone_id
}
