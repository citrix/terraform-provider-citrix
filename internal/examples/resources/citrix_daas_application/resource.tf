resource "citrix_daas_application" "example-application" {
  name                    = "example-name"
  description             = "example-description"
  published_name          = "example-published-name"
  application_folder_path = citrix_daas_application_folder.example-application-folder-1.path
  installed_app_properties = {
    command_line_arguments  = "<Command line arguments for the executable>"
    command_line_executable = "<Command line executable path>"
    working_directory       = "<Working directory path for the executable>"
  }
  delivery_groups = [citrix_daas_delivery_group.example-delivery-group.id]
}
