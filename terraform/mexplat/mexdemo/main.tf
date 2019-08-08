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
    key                   = "mexdemo.tfstate"
  }
}

module "k8s_kr" {
  source              = "../../modules/k8s_azure"

  location            = "${var.azure_kr_location}"
  client_id           = "${var.azure_terraform_service_principal_id}"
  client_secret       = "${var.azure_terraform_service_principal_secret}"
  cluster_name        = "${var.kr_cluster_name}"
  dns_prefix          = "mexdemo-kr"
  vm_size             = "${var.azure_vm_size}"
  cluster_tag         = "mexplat-${var.environ_tag}"
  resource_group_name = "${var.kr_resource_group_name}"
  ssh_public_key      = "${var.ssh_public_key_file}"
}

module "k8s_eu" {
  source              = "../../modules/k8s_azure"

  location            = "${var.azure_eu_location}"
  client_id           = "${var.azure_terraform_service_principal_id}"
  client_secret       = "${var.azure_terraform_service_principal_secret}"
  cluster_name        = "${var.eu_cluster_name}"
  dns_prefix          = "mexdemo-eu"
  vm_size             = "${var.azure_vm_size}"
  cluster_tag         = "mexplat-${var.environ_tag}"
  resource_group_name = "${var.eu_resource_group_name}"
  ssh_public_key      = "${var.ssh_public_key_file}"
}
