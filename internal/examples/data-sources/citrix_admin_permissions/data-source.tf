# Get all predefined admin permissions
data "citrix_admin_permissions" "all-permissions" { }

# The permissions can then be viewed via:
# terraform apply
# terraform state show data.citrix_admin_permissions.all-permissions

# Example using the data source to create a new custom role
data "citrix_admin_role" "desktop_group_admin_role" {
    name = "Desktop Group Custom Admin Role"
    description = "Role for managing delivery groups"
    permissions = [
        // all permissions from the data source that start with "DesktopGroup_"
        for permission in data.citrix_admin_permissions.all-permissions.permissions :
        permission.id if startswith(permission.id, "DesktopGroup_")
    ]
}