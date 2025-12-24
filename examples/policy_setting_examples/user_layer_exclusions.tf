resource "citrix_policy_setting" "user_layer_exclusions" {
    name = "UplUserExclusions"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "C:\\Program Files\\AntiVirusHome\\",
        "C:\\ProgramData\\AntiVirus\\virusdefs.db"
    ])
}
