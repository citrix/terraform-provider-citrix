# Get Admin Scope resource by name
data "citrix_admin_role" "example_admin_role" {
    name = "ExampleAdminRole"
}

# Get Admin Scope resource by id
data "citrix_admin_role" "example_admin_role" {
    id = "00000000-0000-0000-0000-000000000000"
}