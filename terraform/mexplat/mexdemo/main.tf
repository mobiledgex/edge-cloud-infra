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

module "k8s" {
  source              = "../../modules/k8s_azure"

  location            = "${var.azure_location}"
  client_id           = "${var.azure_terraform_service_principal_id}"
  client_secret       = "${var.azure_terraform_service_principal_secret}"
  cluster_name        = "${var.cluster_name}"
  dns_prefix          = "mexdemo2-c-mexdemo2-resourc-902e87"
  vm_size             = "${var.azure_vm_size}"
  cluster_tag         = "mexplat-${var.environ_tag}"
  resource_group_name = "${var.resource_group_name}"
  admin_username      = "azureuser"
  agent_pool_name     = "nodepool1"
  client_id           = "4233042d-4655-4c88-a25f-9cd160e6f16a"
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
}
