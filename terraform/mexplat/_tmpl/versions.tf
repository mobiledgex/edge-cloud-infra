terraform {
  required_version = ">= 0.13"
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
      version = "~> 3.0.2"
    }
    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 3.11.0"
    }
    google = {
      source = "hashicorp/google"
      version = "~> 4.15.0"
    }
  }
}
