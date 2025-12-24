resource "citrix_policy_setting" "citrix_cloud_connectors" {
    name = "WemCloudConnectorList"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://connector1.citrix.com",
        "https://connector2.citrix.com"
    ])
}
