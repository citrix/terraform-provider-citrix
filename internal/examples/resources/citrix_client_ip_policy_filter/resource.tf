resource "citrix_client_ip_policy_filter" "example_client_ip_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled    = true
    allowed    = true
    ip_address = "{IP address to be filtered}"
}