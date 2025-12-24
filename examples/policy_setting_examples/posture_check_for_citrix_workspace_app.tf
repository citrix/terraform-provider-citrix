resource "citrix_policy_setting" "posture_check_for_citrix_workspace_app" {
    name = "AppProtectionPostureCheck"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "AntiKeyLogging",
        "AntiScreenCapture"
    ])
}
