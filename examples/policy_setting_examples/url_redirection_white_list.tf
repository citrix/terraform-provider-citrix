resource "citrix_policy_setting" "url_redirection_white_list" {
    name = "URLRedirectionWhiteList"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://allowed-url1.com",
        "https://allowed-url2.com"
    ])
}
