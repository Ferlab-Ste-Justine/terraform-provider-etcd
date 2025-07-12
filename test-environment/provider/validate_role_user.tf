data "etcd_prefix_range_end" "test2" {
    key = "/test2"
}

data "etcd_prefix_range_end" "testmore2" {
    key = "/testmore2"
}

resource "etcd_role" "test" {
    name = "test"

    permissions {
        permission = "read"
        key = data.etcd_prefix_range_end.test2.key
        range_end = data.etcd_prefix_range_end.test2.range_end
    }
}

resource "etcd_role" "testmore" {
    name = "testmore"

    permissions {
        permission = "read"
        key = data.etcd_prefix_range_end.testmore2.key
        range_end = data.etcd_prefix_range_end.testmore2.range_end
    }
}

resource "etcd_user" "test" {
    username = "test"
    password = "hello"
    roles = ["test", "testmore"]
}

resource "etcd_user" "test2" {
    username = "test2"
    password = "hello"
    roles = []
}