variable "kubernetes_context" {
  description = "Kubernetes context to use in the config file"
  type = string
  default = "microk8s"
}

variable "kubernetes_config" {
  description = "Kubernetes config file to use"
  type = string
  default = "~/.kube/config"
}

variable "kubernetes_resources_prefix" {
  description = "Kubernetes config file to use"
  type = string
  default = "terraform-provider-etcd-"
}

variable "etcd_localhost_port" {
  description = "Etcd localhost port"
  type = number
  default = 32379
}

variable "skip_tls" {
  description = "Skip tls or not"
  type = bool
  default = false
}