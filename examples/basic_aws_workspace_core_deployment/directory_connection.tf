resource "citrix_quickcreate_aws_workspaces_directory_connection" "example_aws_workspaces_directory_connection" {
    name                                = var.directory_connection_name
    account                             = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    resource_location                   = citrix_cloud_resource_location.example_resource_location.id
    directory                           = var.directory_connection_id
    subnets                             = [var.directory_connection_subnet_1, var.directory_connection_subnet_2]
    tenancy                             = var.directory_connection_tenancy
    security_group                      = var.directory_connection_security_group_id
    default_ou                          = var.directory_connection_ou
    user_enabled_as_local_administrator = var.directory_connection_user_as_local_admin
}
