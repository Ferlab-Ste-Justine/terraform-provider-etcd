resource "etcd_key" "hello_world" {
    key = "/hello"
    value = "world"
}