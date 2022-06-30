resource "etcd_synchronized_key_prefixes" "sync_my_app" {
    source_prefix = "/my-app-state-with-bad-prefix/"
    destination_prefix = "/my-app-state-with-the-prefix-i-want/"
    recurrence = "onchange"
}