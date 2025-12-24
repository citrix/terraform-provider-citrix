resource "citrix_policy_setting" "excluded_groups" {
    name = "ExcludedGroups_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "DOMAIN\\ExcludedGroup1",
        "DOMAIN\\ExcludedGroup2"
    ])
}
