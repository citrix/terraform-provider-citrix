# Get Policy data source by id
data "citrix_policy" "example_policy_data_source_with_id" {
    id = "00000000-0000-0000-0000-000000000000"
}

# Get Policy data source by name and policy set id
data "citrix_policy" "example_policy_data_source_with_name" {
    policy_set_id = "00000000-0000-0000-0000-000000000000"
    name = "example-policy-set"
}
