variable "kubernetes_context" {
  description = "Kubernetes context to use in the config file"
  type = string
  default = "minikube"
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
  description = "etcd localhost port"
  type = number
  default = 32379
}