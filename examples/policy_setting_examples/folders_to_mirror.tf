resource "citrix_policy_setting" "folders_to_mirror" {
    name = "MirrorFoldersList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "AppData\\Local\\Microsoft\\Windows\\Cookies"
    ])
}
