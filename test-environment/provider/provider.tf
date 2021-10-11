provider "etcd" {
  endpoints = "127.0.0.1:32379"
  ca_cert = "${path.module}/certs/ca.pem"
  cert = "${path.module}/certs/root.pem"
  key = "${path.module}/certs/root.key"
}