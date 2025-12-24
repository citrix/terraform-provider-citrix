resource "citrix_policy_setting" "replicate_user_stores_paths_to_replicate_a_user_store" {
    name = "MultiSiteReplication_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "\\\\path_a|\\\\path_b",
        "\\\\path_c"
    ])
}
