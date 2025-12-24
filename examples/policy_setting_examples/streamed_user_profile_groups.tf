resource "citrix_policy_setting" "streamed_user_profile_groups" {
    name = "PSUserGroups_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "DOMAIN1\\Group1",
        "DOMAIN2\\Group2"
    ])
}
