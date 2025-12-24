resource "citrix_policy_setting" "exclusion_list" {
    name = "ExclusionList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Software\\Policies"
    ])
}
