provider "azurerm" {
	version = "=1.24.0"

	client_id				= "${var.azure_terraform_service_principal_id}"
	client_secret		= "${var.azure_terraform_service_principal_secret}"
	subscription_id	= "${var.azure_subscription_id}"
	tenant_id				= "${var.azure_tenant_id}"

}

provider "google" {
	version = "=2.5.1"

	project	= "${var.gcp_project}"
	zone		= "${var.gcp_zone}"
}

provider "cloudflare" {
	version	= "=1.14.0"

	email		= "${var.cloudflare_account_email}"
	token		= "${var.cloudflare_account_api_token}"
}

terraform {
	backend "azurerm" {
		storage_account_name	= "mexterraformstate"
		container_name				= "mexplat-tfstate"
		key										= "test.tfstate"
	}
}

module "gitlab" {
	source												= "../../modules/vm_azure"

	instance_name									= "${var.gitlab_instance_name}"
	location											= "${var.azure_location}"
	environ_tag										= "mexplat-${var.environ_tag}"
	resource_group_name						= "mexplat-${var.environ_tag}-rg"
	virtual_network_address_space	= "${var.address_space}"
	subnet_address_prefix					= "${var.address_space}"
}

module "gitlab_dns" {
	source												= "../../modules/cloudflare_record"
	hostname											= "${var.gitlab_domain_name}"
	ip														= "${module.gitlab.external_ip}"
}

module "docker_dns" {
	source												= "../../modules/cloudflare_record"
	hostname											= "${var.gitlab_docker_domain_name}"
	ip														= "${module.gitlab.external_ip}"
}

/*
# VM for console
module "console" {
	source							= "../../modules/vm_gcp"

	instance_name				= "${var.console_instance_name}"
	zone								= "${var.gcp_zone}"
	tags								= [ "http-server", "https-server", "console-debug" ]
}
*/
