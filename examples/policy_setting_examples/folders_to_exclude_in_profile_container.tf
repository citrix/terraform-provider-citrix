resource "citrix_policy_setting" "folders_to_exclude_in_profile_container" {
    name = "ProfileContainerExclusionListDir_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop",
        "Downloads\\*"
    ])
}
