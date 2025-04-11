// Policy Settings are associated with a policy
resource "citrix_policy_setting" "example_policy_setting" {
    policy_id   = citrix_policy.example_policy.id
    name        = "AdvanceWarningPeriod"
    use_default = false
    value       = "13:00:00"
}

// To use default value
resource "citrix_policy_setting" "example_policy_setting" {
    policy_id   = citrix_policy.example_policy.id
    name        = "AdvanceWarningPeriod"
    use_default = true
}