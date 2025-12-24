resource "citrix_policy_setting" "universal_print_servers_for_load_balancing" {
    name = "LoadBalancedPrintServers"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "\\\\printserver1",
        "\\\\printserver2"
    ])
}
