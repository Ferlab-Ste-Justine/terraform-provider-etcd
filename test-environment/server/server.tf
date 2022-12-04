module "etcd_server" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-kubernetes-etcd-localhost.git"
  kubernetes_resources_prefix = var.kubernetes_resources_prefix
  etcd_nodeport = var.etcd_localhost_port
}