locals {
  alertmanager_domain_name = "alertmanager-${var.environ_tag}.${var.dns_domain}"

  console_instance_name = "console-${var.environ_tag}"
  console_domain_name = "${local.console_instance_name}.${var.dns_domain}"
  console_vnc_domain_name = "${local.console_instance_name}-vnc.${var.dns_domain}"

  gitlab_instance_name = "gitlab-${var.environ_tag}"

  harbor_static_address_name = "harbor-${var.environ_tag}"
  harbor_instance_name = "harbor-${var.environ_tag}"
  harbor_domain_name = "${local.harbor_instance_name}.${var.dns_domain}"

  notifyroot_domain_name = "notifyroot-${var.environ_tag}.${var.dns_domain}"

  resource_group_name = "mexplat-${var.environ_tag}-rg"

  stun_domain_name = "stun.${var.dns_domain}"

  vault_a_vm_name = "vault-${var.environ_tag}-a"
  vault_b_vm_name = "vault-${var.environ_tag}-b"
  vault_c_vm_name = "vault-${var.environ_tag}-c"
  vault_a_domain_name = "${local.vault_a_vm_name}.${var.dns_domain}"
  vault_b_domain_name = "${local.vault_b_vm_name}.${var.dns_domain}"
  vault_c_domain_name = "${local.vault_c_vm_name}.${var.dns_domain}"
}
