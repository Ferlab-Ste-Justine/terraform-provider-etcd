//Create some keys in /test3/ prior

data "etcd_prefix_range_end" "test3" {
    key = "/test3/"
}

resource "etcd_range_scoped_state" "test3" {
    key = data.etcd_prefix_range_end.test3.key
    range_end = data.etcd_prefix_range_end.test3.range_end
}