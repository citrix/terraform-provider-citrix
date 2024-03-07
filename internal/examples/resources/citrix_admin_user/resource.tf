resource "citrix_admin_user" "example-admin-user" {
    name = "example-admin-user"
    domain_name = "example-domain"
    rights = [
        {
            role = "Delivery Group Administrator",
            scope = "All"
        }
    ]
    is_enabled = true
}