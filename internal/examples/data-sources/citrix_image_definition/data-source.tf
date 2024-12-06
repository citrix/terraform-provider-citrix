# Define Image Definition data source by id
data "citrix_image_definition" "example_image_definition_by_id" {
    id = "00000000-0000-0000-0000-000000000000"
}

# Define Image Definition data source by name
data "citrix_image_definition" "example_image_definition_by_name" {
    name = "Example Image Definition"
}
