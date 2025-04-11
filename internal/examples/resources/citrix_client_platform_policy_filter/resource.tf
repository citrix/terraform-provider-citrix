resource "citrix_client_platform_policy_filter" "example_client_platform_policy_filter" {
    policy_id   = citrix_policy.example_policy.id
    enabled     = true
    allowed     = true
    platform    = "{Windows|Linux|Mac|Ios|Android|Html5}"
}
