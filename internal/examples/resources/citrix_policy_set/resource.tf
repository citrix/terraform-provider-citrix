resource "citrix_policy_set" "example-policy-set" {
    name = "example-policy-set"
    description = "This is an example policy set description"
    type = "DeliveryGroupPolicies"
    scopes = [ "All", citrix_admin_scope.example-admin-scope.name ]
    policies = [
        {
            name = "test-policy-with-priority-0"
            description = "Test policy in the example policy set with priority 0"
            is_enabled = true
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
            ]
            policy_filters = [
                {
                    type = "DesktopGroup"
                    data = jsonencode({
                        "server" = "20.185.46.142"
                        "uuid" = citrix_policy_set.example-delivery-group.id
                    })
                    is_enabled = true
                    is_allowed = true
                },
            ]
        },
        {
            name = "test-policy-with-priority-1"
            description = "Test policy in the example policy set with priority 1"
            is_enabled = false
            policy_settings = []
            policy_filters = []
        }
    ]
}
