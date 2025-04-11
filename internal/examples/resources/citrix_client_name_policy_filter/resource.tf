resource "citrix_client_name_policy_filter" "example_client_name_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled     = true
    allowed     = true
    client_name = "{Name of the client to be filtered}"
}