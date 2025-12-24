resource "citrix_policy_setting" "onedrive_container_list_of_onedrive_folders" {
    name = "OneDriveContainer_Part"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "OneDrive - Citrix"
    ])
}
