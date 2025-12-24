resource "citrix_policy_setting" "large_file_handling_list_files_to_be_created_as_symbolic_links" {
    name = "LargeFileHandlingList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "!ctx_localappdata!\\Microsoft\\Outlook\\*.OST"
    ])
}
