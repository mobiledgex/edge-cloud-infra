variable "environ_tag" {
  description = "Name to tag instances created by this workspace"
  default     = "test"
}

variable "dns_domain" {
  description = "DNS domain"
  default     = "mobiledgex-dev.net"
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for the DNS domain"
}

variable "cloudflare_account_email" {
  description = "Cloudflare account email"
  type        = string
}

variable "cloudflare_account_api_token" {
  description = "Cloudflare account API token"
  type        = string
}

variable "gcp_project" {
  description = "GCP project to use"
  default = "still-entity-201400"
}

variable "gcp_zone" {
  description = "Default GCP zone for resources"
  default = "us-west2-a"
}

variable "gitlab_gcp_zone" {
  description = "GCP zone for the Gitlab VM"
  default = "us-west1-b"
}

variable "vault_a_gcp_zone" {
  description = "GCP zone for Vault HA instance A"
  default = "us-central1-a"
}

variable "vault_b_gcp_zone" {
  description = "GCP zone for Vault HA instance B"
  default = "europe-west3-a"
}

variable "vault_c_gcp_zone" {
  description = "GCP zone for Vault HA instance C"
  default     = "asia-east1-b"
}

variable "global_tags" {
  default     = [
    "iap-ssh",
    "restricted-ssh",
    "restricted-ssh-overrides",
  ]
}
