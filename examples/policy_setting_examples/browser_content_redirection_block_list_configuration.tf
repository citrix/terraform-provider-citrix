resource "citrix_policy_setting" "browser_content_redirection_block_list_configuration" {
    name = "WebBrowserRedirectionBlacklist"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "https://blocked-site1.com/*",
        "https://blocked-site2.com/*"
    ])
}
