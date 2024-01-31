resource "citrix_daas_admin_role" "on_prem_example_role" {
    name = "on_prem_admin_role"
    description = "Example admin role for citrix onprem"
    permissions = ["AppGroupApplications_ChangeTags"] 
}

resource "citrix_daas_admin_role" "cloud_example_role" {
    name = "cloud_admin_role"
    can_launch_manage = false
    can_launch_monitor = true
    description = "Example admin role for citrix daas"
    permissions = [
        "AppGroupApplications_Read", 
        "ApplicationGroup_AddScope", 
        "ApplicationGroup_Read"
    ]
}