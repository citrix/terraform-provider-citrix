// Available `delivery_group_type` are `Private`, `PrivateApp`, `Shared`, `SharedApp`
// `Private` means Private Desktop
// `PrivateApp` means Private Application
// `Shared` means Shared Desktop
// `SharedApp` means Shared Application
resource "citrix_delivery_group_type_policy_filter" "example_delivery_group_type_policy_filter" {
    policy_id           = citrix_policy.example_policy.id
    enabled             = true
    allowed             = true
    delivery_group_type = "{Type of the delivery group to be filtered}"
}