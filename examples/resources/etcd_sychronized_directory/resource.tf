//Upload some prometheus configs
resource "etcd_synchronized_directory" "prometheus_confs" {
    provider = etcdnew
    directory = "${path.module}/prometheus-confs"
    key_prefix = "/prometheus-confs/"
    source = "directory"
    recurrence = "once"
}

//sync key range in etcdnew with the one in etcdold
resource "etcd_synchronized_directory" "source" {
    provider = etcdold
    directory = "/tmp/prefix-to-transfer"
    key_prefix = "/prefix-to-transfer/"
    source = "key-prefix"
    recurrence = "once"
}

resource "etcd_synchronized_directory" "destination" {
    provider = etcdnew
    directory = "/tmp/prefix-to-transfer"
    key_prefix = "/prefix-to-transfer/"
    source = "directory"
    recurrence = "once"

    depends_on = [etcd_synchronized_directory.source]
}