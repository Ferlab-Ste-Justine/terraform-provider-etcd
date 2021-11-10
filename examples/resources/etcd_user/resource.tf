resource "etcd_role" "summary_reader" {
    name = "summary"

    permissions {
        permission = "read"
        key = "/summary.txt"
        range_end = "/summary.txt"
    }
}

resource "etcd_user" "bob" {
    username = "bob"
    password = "1234"
    roles = [etcd_role.summary_reader.name]
}