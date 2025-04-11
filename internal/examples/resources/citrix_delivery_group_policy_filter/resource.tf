resource "citrix_delivery_group_policy_filter" "example_delivery_group_policy_filter" {
    policy_id         = citrix_policy.example_policy.id
    enabled           = true
    allowed           = true
    delivery_group_id = "{ID of the delivery group to be filtered}"
}