resource "citrix_policy_setting" "profile_streaming_exclusion_list_directories" {
    name = "StreamingExclusionList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Desktop"
    ])
}
