output "kube_config" {
  value = "${module.k8s.kube_config}"
}

output "gitlab_external_ip" {
  value = "${module.gitlab.external_ip}"
}

# Same as the Gitlab VM
output "crm_external_ip" {
  value = "${module.gitlab.external_ip}"
}

# Same as the Gitlab VM
output "mc_external_ip" {
  value = "${module.gitlab.external_ip}"
}

# Same as the Gitlab VM
output "postgres_external_ip" {
  value = "${module.gitlab.external_ip}"
}

# Same as the Gitlab VM
output "vault_external_ip" {
  value = "${module.gitlab.external_ip}"
}

output "console_external_ip" {
  value = "${module.console.external_ip}"
}

output "k8s_cluster_name" {
  value = "${var.cluster_name}"
}

output "eu_kube_config" {
  value = "${module.k8s_eu.kube_config}"
}

output "eu_k8s_cluster_name" {
  value = "${var.eu_cluster_name}"
}

output "k8s_clusters" {
  value = [
    {
      "name" = "${var.cluster_name}"
      "kube_config" = "${module.k8s.kube_config}"
      "region" = "US"
    },
    {
      "name" = "${var.eu_cluster_name}"
      "kube_config" = "${module.k8s_eu.kube_config}"
      "region" = "EU"
    }
  ]
}
