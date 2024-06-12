resource "citrix_policy_set" "rv2-policy-set" {
    name = "rv2-policy-set"
    description = "This is an policy set to enable rendezvous v2 for Connectorless VDAs"
    type = "DeliveryGroupPolicies"
    policies = [
        {
            name = "rv2-policy-with-priority-0"
            enabled = true
            policy_settings = [
                {
                    name = "RendezvousProtocol"
                    enabled = true
                    use_default = false
                },
            ]
            delivery_group_filters = [
                {
                    delivery_group_id = citrix_delivery_group.example-delivery-group.id
                    enabled = true
                    allowed = true
                },
            ]
        }
    ]
}