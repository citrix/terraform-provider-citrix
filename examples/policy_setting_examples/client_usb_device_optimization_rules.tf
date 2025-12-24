resource "citrix_policy_setting" "client_usb_device_optimization_rules" {
    name = "ClientUsbDeviceOptimizationRules"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Mode=00000004 VID=1230 PID=1230 class=03 #Input device operating in capture mode",
        "Mode=00000002 VID=1230 PID=1230 class=03 #Input device operating in interactive mode"
    ])
}
