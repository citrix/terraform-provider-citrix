resource "citrix_application_folder" "example-application-folder-1" {
  name               = "example-application-folder-1"
}

resource "citrix_application_folder" "example-application-folder-2" {
  name               = "example-application-folder-2"
  parent_path        = citrix_application_folder.example-application-folder-1.path
}
