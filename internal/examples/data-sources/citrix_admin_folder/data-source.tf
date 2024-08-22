# Get Admin Folder resource by id
data "citrix_admin_folder" "test_admin_folder_data_source_with_id" {
    id = "0"
}

# Get Admin Folder resource by path
data "citrix_admin_folder" "test_admin_folder_data_source_with_path" {
    path = "test123\\"
}
