resource "citrix_user_policy_filter" "example_user_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled     = true
    allowed     = true
    sid         = "{SID of the user or user group to be filtered}"
}