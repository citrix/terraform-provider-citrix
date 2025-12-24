resource "citrix_policy_setting" "folders_to_include_in_profile_container" {
    name = "ProfileContainerInclusionListDir_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop\\include",
        "Downloads\\include"
    ])
}
