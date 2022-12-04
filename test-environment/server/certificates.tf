resource "local_file" "ca_cert" {
  content = module.etcd_server.ca_certificate
  filename = "${path.module}/certs/ca.pem"
  file_permission = "0600"
}

resource "local_file" "server_cert" {
  content = module.etcd_server.server_certificate
  filename = "${path.module}/certs/server.pem"
  file_permission = "0600"
}

resource "local_file" "server_key" {
  content = module.etcd_server.server_key
  filename = "${path.module}/certs/server.key"
  file_permission = "0600"
}

resource "local_file" "root_cert" {
  content = module.etcd_server.root_certificate
  filename = "${path.module}/certs/root.pem"
  file_permission = "0600"
}

resource "local_file" "root_key" {
  content = module.etcd_server.root_key
  filename = "${path.module}/certs/root.key"
  file_permission = "0600"
}