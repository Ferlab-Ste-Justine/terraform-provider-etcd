provider "kubernetes" {
  config_path    = var.kubernetes_config
  config_context = var.kubernetes_context
}
