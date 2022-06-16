data "etcd_prefix_range_end" "patroni" {
  key = "/patroni/"
}

resource "etcd_range_scoped_state" "patroni" {
    key = data.etcd_prefix_range_end.patroni.key
    range_end = data.etcd_prefix_range_end.patroni.range_end
    clear_on_creation = false
    clear_on_deletion = true
}