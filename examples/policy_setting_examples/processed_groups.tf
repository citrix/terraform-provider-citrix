resource "citrix_policy_setting" "processed_groups" {
    name = "ProcessedGroups_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "DOMAIN\\GROUPNAME1",
        "DOMAIN\\GROUPNAME2"
    ])
}
