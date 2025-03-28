resource "citrix_admin_role" "on_premises_example_role" {
    name = "on_premise_admin_role"
    description = "Example admin role for CVAD onpremises"
    permissions = ["AppGroupApplications_ChangeTags"] 
}

resource "citrix_admin_role" "cloud_example_role" {
    name = "cloud_admin_role"
    can_launch_manage = true
    can_launch_monitor = false
    description = "Example admin role for a DaaS admin with some app group permissions"
    permissions = [
        "AppGroupApplications_Read", 
        "ApplicationGroup_AddScope", 
        "ApplicationGroup_Read"
    ]
}

# To view all possible permissions use the data source:
data "citrix_admin_permissions" "all-permissions" { }

# The permissions can then be viewed via:
# terraform apply
# terraform state show data.citrix_admin_permissions.all-permissions

# Example using the data source to create a new custom role
data "citrix_admin_role" "desktop_group_admin_role" {
    name = "Delivery Group Custom Admin Role"
    description = "Role for managing delivery groups"
    permissions = [
        // all permissions from the data source that start with "DesktopGroup_", the old name for delivery groups
        for permission in data.citrix_admin_permissions.all-permissions.permissions :
        permission.id if startswith(permission.id, "DesktopGroup_")
    ]
}