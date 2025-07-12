resource "etcd_key" "key_to_read" {
    key = "/key/to/read"
    value = "hello"
    clear_on_deletion = false
}

data "etcd_key" "key_to_read" {
    key = "/key/to/read"
    must_exist = false
}

output "key_to_read" {
  value     = data.etcd_key.key_to_read
}

data "etcd_key" "non_existing_read" {
    key = "/does/not/exist"
    must_exist = false
}

output "non_existing_read" {
  value     = data.etcd_key.non_existing_read
}