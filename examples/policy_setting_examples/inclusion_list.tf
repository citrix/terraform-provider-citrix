resource "citrix_policy_setting" "inclusion_list" {
    name = "IncludeListRegistry_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Software\\Adobe"
    ])
}
