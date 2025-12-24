resource "citrix_policy_setting" "set_priority_order_for_user_groups" {
    name = "OrderedGroups_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = "ctxxa.local\\groupb;S-1-5-21-674278408-26188528-2146851469-1174;ctxxa.local\\groupc"
}
