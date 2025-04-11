// With Citrix Gateway
resource "citrix_access_control_policy_filter" "example_access_control_policy_filter" {
    policy_id       = citrix_policy.example_policy.id
    enabled         = true
    allowed         = true
    connection_type = "WithAccessGateway"
    condition       = "{Access Condition}" // Wildcard `*` is allowed
    gateway         = "{Gateway farm name}" // Wildcard `*` is allowed
}

// Without Citrix Gateway
resource "citrix_access_control_policy_filter" "example_access_control_policy_filter" {
    policy_id       = citrix_policy.example_policy.id
    enabled         = true
    allowed         = true
    connection_type = "WithoutAccessGateway"
    condition       = "*"
    gateway         = "*"
}