# Get Tag detail by name
data "citrix_tag" "example_tag_by_name" {
    name = "exampleTag"
}

# Get Tag detail by id
data "citrix_tag" "example_tag_by_id" {
    id = "00000000-0000-0000-0000-000000000000"
}