resource "citrix_admin_folder" "example-admin-folder-1" {
  name          = "example-admin-folder-1"
  type          = "ContainsApplications"
}

resource "citrix_admin_folder" "example-admin-folder-2" {
  name          = "example-admin-folder-2"
  type          = "ContainsApplications"
  parent_path   = citrix_admin_folder.example-admin-folder-1.path
}
