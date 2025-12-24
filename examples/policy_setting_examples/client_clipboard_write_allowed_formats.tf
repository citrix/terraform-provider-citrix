resource "citrix_policy_setting" "client_clipboard_write_allowed_formats" {
    name = "ClientClipboardWriteAllowedFormats"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "CF_TEXT",
        "CFX_FILE"
    ])
}
