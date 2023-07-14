provider "etcd" {
  endpoints = "127.0.0.1:32379"
  ca_cert = var.skip_tls ? null : "${path.module}/../server/certs/ca.pem"
  cert = var.skip_tls ? null : "${path.module}/../server/certs/root.pem"
  key = var.skip_tls ? null : "${path.module}/../server/certs/root.key"
  username = var.skip_tls ? "root" : null
  password = var.skip_tls ? file("${path.module}/../server/certs/root_password") : null
  skip_tls = var.skip_tls
}