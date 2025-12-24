resource "citrix_policy_setting" "files_to_exclude_in_profile_container" {
    name = "ProfileContainerExclusionListFile_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop\\Desktop.ini",
        "AppData\\*.tmp"
    ])
}
