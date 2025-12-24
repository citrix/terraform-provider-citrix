resource "citrix_policy_setting" "directories_to_synchronize" {
    name = "SyncDirList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop\\exclude\\include",
        "Desktop\\exclude\\*"
    ])
}
