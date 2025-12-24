resource "citrix_policy_setting" "files_to_exclude_from_the_shared_store" {
    name = "SharedStoreFileExclusionList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Downloads\\profilemgt_x64.msi",
        "*.tmp"
    ])
}
