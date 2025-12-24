resource "citrix_policy_setting" "session_clipboard_write_allowed_formats" {
    name = "SessionClipboardWriteAllowedFormats"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "CF_TEXT",
        "CFX_FILE"
    ])
}
