resource "citrix_policy_setting" "virtual_ip_virtual_loopback_programs_list" {
    name = "VirtualLoopbackPrograms"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "C:\\Program Files\\Application1\\app1.exe",
        "C:\\Program Files\\Application2\\app2.exe"
    ])
}
