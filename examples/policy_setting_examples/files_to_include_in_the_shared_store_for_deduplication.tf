resource "citrix_policy_setting" "files_to_include_in_the_shared_store_for_deduplication" {
    name = "SharedStoreFileInclusionList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Downloads\\profilemgt_x64.msi",
        "*.cfg",
        "Music\\*"
    ])
}
