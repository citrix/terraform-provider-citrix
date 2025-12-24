resource "citrix_policy_setting" "browser_content_redirection_acl_configuration" {
    name = "WebBrowserRedirectionAcl"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://www.example.com/*",
        "https://some.website.com/index.html"
    ])
}
