resource "etcd_key_prefix" "hello_world" {
    prefix = "/hello_world/"
    clear_on_deletion = true

    keys {
        key = "hello"
        value = "hello"
    }

    keys {
        key = "world"
        value = "world"
    }
}