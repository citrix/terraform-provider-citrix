resource "citrix_policy_setting" "virtual_channel_allow_list" {
    name = "VirtualChannelWhiteList"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "CTXCVC1,C:\\VC1\\vchost.exe",
        "CTXCVC2,C:\\VC2\\vchost.exe,C:\\Program Files\\Third Party\\vcaccess.exe"
    ])
}
