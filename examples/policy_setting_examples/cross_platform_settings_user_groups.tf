resource "citrix_policy_setting" "cross_platform_settings_user_groups" {
    name = "CPUserGroups_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "DOMAIN\\UserGroup1",
        "DOMAIN\\UserGroup2"
    ])
}
