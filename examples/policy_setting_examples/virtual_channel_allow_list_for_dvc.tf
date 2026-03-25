resource "citrix_policy_setting" "virtual_channel_allow_list_for_dvc" {
    name = "DynamicVirtualChannelAllowList"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "C:\\VC1\\vchost.exe",
        "C:\\VC2\\vchost.exe",
        "C:\\Program Files\\Third Party\\vcaccess.exe"
    ])
}
