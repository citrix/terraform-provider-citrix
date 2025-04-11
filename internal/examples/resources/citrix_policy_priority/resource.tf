// Policy Priority manages the priorities of the policies in a policy set.
resource "citrix_policy_priority" "example" {
    policy_set_id   = citrix_policy_set_v2.example_policy_set_v2.id
    policy_priority = [
        citrix_policy.example_policy_1.id,
        citrix_policy.example_policy_2.id,
        citrix_policy.example_policy_3.id
    ]
}