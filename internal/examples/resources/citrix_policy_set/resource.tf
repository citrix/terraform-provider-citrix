resource "citrix_policy_set" "example-policy-set" {
    name = "example-policy-set"
    description = "This is an example policy set description"
    type = "DeliveryGroupPolicies"
    scopes = [ "All", citrix_admin_scope.example-admin-scope.name ]
    policies = [
        {
            name = "test-policy-with-priority-0"
            description = "Test policy in the example policy set with priority 0"
            enabled = true
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
                    data = {
                        server = "0.0.0.0"
                        uuid = citrix_delivery_group.example-delivery-group.id
                    }
                    enabled = true
                    allowed = true
                },
            ]
        },
        {
            name = "test-policy-with-priority-1"
            description = "Test policy in the example policy set with priority 1"
            enabled = false
            policy_settings = []
            policy_filters = []
        }
    ]
}
