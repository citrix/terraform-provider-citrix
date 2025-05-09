# Application resource with priorities of delivery groups specified with the `delivery_groups` list attribute.
resource "citrix_application" "example-application" {
  name                    = "example-name"
  description             = "example-description"
  published_name          = "example-published-name"
  application_folder_path = citrix_admin_folder.example-admin-folder-for-application.path
  installed_app_properties = {
    command_line_arguments  = "<Command line arguments for the executable>"
    command_line_executable = "<Command line executable path>"
    working_directory       = "<Working directory path for the executable>"
  }
  delivery_groups = [citrix_delivery_group.example-delivery-group.id]
  icon            = citrix_application_icon.example-application-icon.id
  limit_visibility_to_users = ["example\\user1"]
  application_category_path = "ApplicationCategory\\SubCategory"
  shortcut_added_to_desktop = true
  shortcut_added_to_start_menu = false
}

# Application resource with priorities of delivery groups specified with the `delivery_groups_priority` set attribute.
resource "citrix_application" "example-application" {
  name                    = "example-name"
  description             = "example-description"
  published_name          = "example-published-name"
  application_folder_path = citrix_admin_folder.example-admin-folder-for-application.path
  installed_app_properties = {
    command_line_arguments  = "<Command line arguments for the executable>"
    command_line_executable = "<Command line executable path>"
    working_directory       = "<Working directory path for the executable>"
  }
  delivery_groups_priority = [
    {
      id = citrix_delivery_group.example-delivery-group-1.id
      priority = 3
    },
    {
      id = citrix_delivery_group.example-delivery-group-2.id
      priority = 0
    }
  ]
  icon            = citrix_application_icon.example-application-icon.id
  limit_visibility_to_users = ["example\\user1"]
  application_category_path = "ApplicationCategory\\SubCategory"
}

# Application resource with CPU priority high specified with the `cpu_priority_level` attribute.
resource "citrix_application" "example-application" {
  name                    = "example-name"
  description             = "example-description"
  published_name          = "example-published-name"
  application_folder_path = citrix_admin_folder.example-admin-folder-for-application.path
  installed_app_properties = {
    command_line_arguments  = "<Command line arguments for the executable>"
    command_line_executable = "<Command line executable path>"
    working_directory       = "<Working directory path for the executable>"
  }
  delivery_groups = [citrix_delivery_group.example-delivery-group.id]
  icon            = citrix_application_icon.example-application-icon.id
  limit_visibility_to_users = ["example\\user1"]
  cpu_priority_level = "High"
}
