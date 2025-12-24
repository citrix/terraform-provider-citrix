resource "citrix_policy_setting" "allowed_urls_to_be_redirected_to_client" {
    name = "ClientURLs"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://allowed-url1.com",
        "https://*.allowed-url2.com",
        "https://allowed-url3.com"
    ])
}