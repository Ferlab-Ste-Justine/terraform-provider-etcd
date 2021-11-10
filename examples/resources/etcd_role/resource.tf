data "etcd_prefix_range_end" "conf_files" {
    key = "/confs/"
}

resource "etcd_role" "configurator" {
    name = "configurator"

    permissions {
        permission = "readwrite"
        key = data.etcd_prefix_range_end.conf_files.key
        range_end = data.etcd_prefix_range_end.conf_files.range_end
    }

    permissions {
        permission = "read"
        key = "/summary.txt"
        range_end = "/summary.txt"
    }
}