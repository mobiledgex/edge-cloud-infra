variable "image" {
  default = "cirros-0.3.5-x86_64-disk"
}

variable "flavor" {
  default = "m1.tiny"
}

variable "ssh_key_file" {
  default = "~/.ssh/id_rsa.terraform"
}

variable "ssh_user_name" {
  default = "ubuntu"
}

variable "external_gateway" {}

variable "pool" {
  default = "public"
}
