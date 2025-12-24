resource "citrix_policy_setting" "client_usb_device_redirection_rules_version_2" {
    name = "USBDeviceRulesV2"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "ALLOW: vid=1188 pid=A101",
        "DENY: class=09"
    ])
}
