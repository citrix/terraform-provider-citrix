resource "citrix_policy_setting" "exclusion_list_files" {
    name = "ExclusionListSyncFiles_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop\\Desktop.ini",
        "AppData\\*.tmp"
    ])
}
