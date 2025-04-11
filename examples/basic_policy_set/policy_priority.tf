// Example with policy priority first_basic_policy = 0, second_basic_policy = 1, third_basic_policy = 2
resource "citrix_policy_priority" "example_policy_priority" {
    policy_set_id    = citrix_policy_set_v2.basic_policy_set.id
    policy_priority  = [
        citrix_policy.first_basic_policy.id,
        citrix_policy.second_basic_policy.id,
        citrix_policy.third_basic_policy.id
    ]
}
