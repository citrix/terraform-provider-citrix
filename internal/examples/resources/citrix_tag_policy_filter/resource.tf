resource "citrix_tag_policy_filter" "example_tag_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled     = true
    allowed     = true
    tag         = "{ID of the tag to be filtered}"
}