// Policy Settings are depending on the `citrix_policy` resource.
// Since the `citrix_policy` resource depends on `citrix_policy_set_v2` resource, the `citrix_policy_setting` resource has an implicit dependency on the `citrix_policy_set_v2` resource.
resource "citrix_policy_setting" "advance_warning_period" {
    policy_id   = citrix_policy.first_basic_policy.id
    name        = "AdvanceWarningPeriod"
    use_default = false
    value       = "13:00:00"
}

resource "citrix_policy_setting" "visually_lossless_compression" {
    policy_id   = citrix_policy.first_basic_policy.id
    name        = "AllowVisuallyLosslessCompression"
    use_default = false
    enabled     = true
}

resource "citrix_policy_setting" "wem_citrix_cloud_connectors" {
    policy_id   = citrix_policy.first_basic_policy.id
    name        = "WemCloudConnectorList"
    use_default = false
    value       = jsonencode([
        "https://connector1.citrix.com",
        "https://connector2.citrix.com"
    ])
}
