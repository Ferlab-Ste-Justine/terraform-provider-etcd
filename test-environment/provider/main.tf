resource "etcd_role" "test" {
    name = "test"

    permissions {
        permission = "readwrite"
        key = "/test"
        range_end = "/test"
    }

    permissions {
        permission = "read"
        key = "/test2"
        range_end = "/test3"
    }
}

resource "etcd_role" "testmore" {
    name = "testmore"

    permissions {
        permission = "readwrite"
        key = "/testmore"
        range_end = "/testmore"
    }

    permissions {
        permission = "read"
        key = "/testmore2"
        range_end = "/testmore3"
    }
}

resource "etcd_user" "test" {
    username = "test"
    password = "hello"
    roles = ["test", "testmore"]
}

resource "etcd_key" "test" {
    key = "/test/hello"
    value = "worldy"
}