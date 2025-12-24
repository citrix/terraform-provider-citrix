resource "citrix_policy_setting" "browser_content_redirection_authentication_sites" {
    name = "WebBrowserRedirectionAuthenticationSites"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://www.example.com/*",
        "https://some.website.com/index.html"
    ])
}
