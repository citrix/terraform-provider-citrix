resource "citrix_policy_setting" "files_to_include_in_profile_container" {
    name = "ProfileContainerInclusionListFile_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "AppData\\Local\\Microsoft\\Office\\Access.qat",
        "AppData\\Local\\MyApp\\*.cfg"
    ])
}
