resource "citrix_policy_setting" "universal_driver_preference" {
    name = "UniversalDriverPriority"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Citrix Universal Printer",
        "Citrix XPS Universal Printer"
    ])
}
