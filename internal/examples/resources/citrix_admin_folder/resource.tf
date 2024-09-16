resource "citrix_admin_folder" "example-admin-folder-1" {
  name          = "example-admin-folder-1"
  type          = [ "ContainsApplications" ]
}

resource "citrix_admin_folder" "example-admin-folder-2" {
  name          = "example-admin-folder-2"
  type          = [ "ContainsApplications" ]
  parent_path   = citrix_admin_folder.example-admin-folder-1.path
}

# If you want to define admin folders with different types but with the same name, please use a single resource block with a set of types for the admin folder
resource "citrix_admin_folder" "example-admin-folder-3" {
  name          = "example-admin-folder-3"
  type          = [ "ContainsApplications", "ContainsMachineCatalogs", "ContainsDeliveryGroups", "ContainsApplicationGroups" ]
}
