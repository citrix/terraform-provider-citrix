# Get Policy Set data source by id
data "citrix_policy_set" "example_policy_set_data_source_with_id" {
    id = "00000000-0000-0000-0000-000000000000"
}

# Get Policy Set data source by name
data "citrix_policy_set" "example_policy_set_data_source_with_name" {
    name = "example-policy-set"
}