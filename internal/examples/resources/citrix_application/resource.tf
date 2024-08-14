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
}
