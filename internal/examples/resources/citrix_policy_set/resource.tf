resource "citrix_policy_set" "example-policy-set" {
    name = "example-policy-set"
    description = "This is an example policy set description"
    type = "DeliveryGroupPolicies"
    scopes = [ citrix_admin_scope.example-admin-scope.id ]
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
            access_control_filters = [
                {
                    connection = "WithAccessGateway"
                    condition  = "*"
                    gateway    = "*"
                    enabled    = true
                    allowed    = true
                },
            ]
            branch_repeater_filter = {
                enabled = true
                allowed = true
            },
            client_ip_filters = [
                {
                    ip_address = "10.0.0.1"
                    enabled    = true
                    allowed    = true
                }
            ]
            client_name_filters = [
                {
                    client_name = "Example Client Name"
                    enabled     = true
                    allowed     = true
                }
            ]
            delivery_group_filters = [
                {
                    delivery_group_id = citrix_delivery_group.example-delivery-group.id
                    enabled           = true
                    allowed           = true
                },
            ]
            delivery_group_type_filters = [
                {
                    delivery_group_type = "Private"
                    enabled             = true
                    allowed             = true
                },
            ]
            ou_filters = [
                {
                    ou     = "{Path of the oranizational unit to be filtered}"
                    enabled = true
                    allowed = true
                },
            ]
            user_filters = [
                {
                    sid     = "{SID of the user or user group to be filtered}"
                    enabled = true
                    allowed = true
                },
            ]
            tag_filters = [
                {
                    tag     = "{ID of the tag to be filtered}"
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
        }
    ]
}
