resource "citrix_policy_setting" "groups_using_customized_user_layer_size" {
    name = "UplGroupsUsingCustomizedUserLayerSize"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "DOMAIN\\Group1",
        "DOMAIN\\Group2"
    ])
}
