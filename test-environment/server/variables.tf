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