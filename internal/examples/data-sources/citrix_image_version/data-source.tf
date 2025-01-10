# Define Image Version data source by id
data "citrix_image_version" "example_image_version_by_id" {
    id               = "00000000-0000-0000-0000-000000000000"
    image_definition = "00000000-0000-0000-0000-000000000000"
}

# Define Image Version data source by version number
data "citrix_image_version" "example_image_version_by_version_number" {
    version_number = 1
    image_definition = "00000000-0000-0000-0000-000000000000"
}
