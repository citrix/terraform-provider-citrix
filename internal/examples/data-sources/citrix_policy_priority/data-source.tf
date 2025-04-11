# Get Policy Priority data source by policy_set_id
data "citrix_policy_priority" "example_policy_priority_data_source_with_policy_set_id" {
    policy_set_id = "00000000-0000-0000-0000-000000000000"
}

# Get Policy Priority data source by policy_set_name
data "citrix_policy" "example_policy_data_source_with_name" {
    policy_set_name = "example-policy-set"
}
