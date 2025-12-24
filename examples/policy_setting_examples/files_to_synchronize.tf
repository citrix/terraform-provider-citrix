resource "citrix_policy_setting" "files_to_synchronize" {
    name = "SyncFileList_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "AppData\\Local\\Microsoft\\Office\\Access.qat",
        "AppData\\Local\\MyApp\\*.cfg"
    ])
}
