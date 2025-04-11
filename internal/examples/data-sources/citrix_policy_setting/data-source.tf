# Get Policy Setting data source by id
data "citrix_policy_setting" "example_policy_setting_data_source_with_id" {
    id = "00000000-0000-0000-0000-000000000000"
}

# Get Policy data source by name and policy set id
data "citrix_policy_setting" "example_policy_setting_data_source_with_name" {
    policy_id = "00000000-0000-0000-0000-000000000000"
    name      = "AdvanceWarningPeriod"
}
