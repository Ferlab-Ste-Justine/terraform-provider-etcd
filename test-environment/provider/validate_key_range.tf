resource "etcd_key" "test" {
    key = "/test/hello"
    value = "worldy"
}

resource "etcd_key" "test2" {
    key = "/test/hello2"
    value = "worldy2"
}

resource "etcd_key" "test3" {
    key = "/test/hello3"
    value = "worldy3"
}

data "etcd_prefix_range_end" "test" {
    key = "/test/"
}

data "etcd_key_range" "test_range" {
    key = data.etcd_prefix_range_end.test.key
    range_end = data.etcd_prefix_range_end.test.range_end
    depends_on = [
        etcd_key.test,
        etcd_key.test2,
        etcd_key.test3
    ]
}

output "test_range" {
  value     = data.etcd_key_range.test_range
}