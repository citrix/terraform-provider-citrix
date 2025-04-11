# When `allowed` is set to `true`, this means policy is applied to `Connections with Citrix SD-WAN`. 
resource "citrix_branch_repeater_policy_filter" "example_branch_repeater_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    allowed     = true
}

# When `allowed` is set to `false`, this means policy is applied to `Connections without Citrix SD-WAN`.
resource "citrix_branch_repeater_policy_filter" "example_branch_repeater_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    allowed     = false
}