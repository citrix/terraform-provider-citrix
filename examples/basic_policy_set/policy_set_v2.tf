resource "citrix_policy_set_v2" "basic_policy_set" {
    name        = "basic policy set"
    description = "basic policy set description"
    scopes      = [
        var.scope_id1,
        var.scope_id2
    ]
    delivery_groups = [
        var.delivery_group_id1,
        var.delivery_group_id2
    ]
}
