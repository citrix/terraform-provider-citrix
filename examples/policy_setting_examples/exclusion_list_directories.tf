resource "citrix_policy_setting" "exclusion_list_directories" {
    name = "ExclusionListSyncDir_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop",
        "Downloads\\*"
    ])
}
