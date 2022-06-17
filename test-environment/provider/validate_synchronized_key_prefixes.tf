resource "etcd_key" "tocopy1" {
    key = "/to-copy/copy1"
    value = "copy1"
}

resource "etcd_key" "tocopy2" {
    key = "/to-copy/copy2"
    value = "copy2"
}

resource "etcd_key" "tocopy3" {
    key = "/to-copy/copy3"
    value = "copy3"
}

resource "etcd_synchronized_key_prefixes" "copy" {
    source_prefix = "/to-copy/"
    destination_prefix = "/copy/"
    recurrence = "always"

    depends_on = [
        etcd_key.tocopy1,
        etcd_key.tocopy2,
        etcd_key.tocopy3,
    ]
}

data "etcd_prefix_range_end" "tocopy" {
    key = "/to-copy/"
}

data "etcd_key_range" "tocopy" {
    key = data.etcd_prefix_range_end.tocopy.key
    range_end = data.etcd_prefix_range_end.tocopy.range_end
    depends_on = [
        etcd_key.tocopy1,
        etcd_key.tocopy2,
        etcd_key.tocopy3,
    ]
}

data "etcd_prefix_range_end" "copy" {
    key = "/copy/"
}

data "etcd_key_range" "copy" {
    key = data.etcd_prefix_range_end.copy.key
    range_end = data.etcd_prefix_range_end.copy.range_end
    depends_on = [
        etcd_synchronized_key_prefixes.copy
    ]
}

output "tocopy" {
  value     = data.etcd_key_range.tocopy
}

output "copy" {
  value     = data.etcd_key_range.copy
}