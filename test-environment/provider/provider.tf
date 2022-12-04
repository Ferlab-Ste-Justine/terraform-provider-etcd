provider "etcd" {
  endpoints = "127.0.0.1:32379"
  ca_cert = "${path.module}/../server/certs/ca.pem"
  cert = "${path.module}/../server/certs/root.pem"
  key = "${path.module}/../server/certs/root.key"
}