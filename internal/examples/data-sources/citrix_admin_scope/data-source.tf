# Get Admin Scope resource by name
data "citrix_admin_scope" "test_scope_by_name" {
    name = "All"
}

# Get Admin Scope resource by id
data "citrix_admin_scope" "test_scope_by_id" {
    id = "00000000-0000-0000-0000-000000000000"
}