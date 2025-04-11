resource "citrix_ou_policy_filter" "example_ou_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled     = true
    allowed     = true
    ou          = "{Path of the organizational unit to be filtered}"
}