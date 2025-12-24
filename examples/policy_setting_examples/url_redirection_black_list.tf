resource "citrix_policy_setting" "url_redirection_black_list" {
    name = "URLRedirectionBlackList"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://blocked-url1.com",
        "https://blocked-url2.com"
    ])
}
