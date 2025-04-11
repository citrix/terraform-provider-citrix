// Policy set is a collection of policies. You can use `citrix_policy_priority` resource to set the priority of the policies in the policy set.
resource "citrix_policy_set_v2" "example_policy_set_v2" {
    name        = "example_policy_set_v2"
    description = "example_policy_set_v2 description"
    scopes      = []
    delivery_groups = [
        "00000000-0000-0000-0000-000000000000"
    ]
}

// The default policy set can also be managed by the `citrix_policy_set_v2` resource by import. However, it cannot be modified or deleted.
resource "citrix_policy_set_v2" "default_site_policies" {
    name        = "DefaultSitePolicies"
    delivery_groups = []
}