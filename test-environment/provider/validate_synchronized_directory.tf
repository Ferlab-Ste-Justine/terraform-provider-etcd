resource "etcd_synchronized_directory" "source" {
    directory = "${path.module}/dir-sync"
    key_prefix = "/dir-sync/"
    source = "directory"
    recurrence = "always"
}

resource "etcd_synchronized_directory" "destination" {
    directory = "${path.module}/dir-sync-copy"
    key_prefix = "/dir-sync/"
    source = "key-prefix"
    recurrence = "always"

    depends_on = [etcd_synchronized_directory.source]
}

data "etcd_prefix_range_end" "dir_sync" {
    key = "/dir-sync/"
}

data "etcd_key_range" "dir_sync" {
    key = data.etcd_prefix_range_end.dir_sync.key
    range_end = data.etcd_prefix_range_end.dir_sync.range_end
    depends_on = [etcd_synchronized_directory.source]
}

output "dir_sync" {
  value     = data.etcd_key_range.dir_sync
}