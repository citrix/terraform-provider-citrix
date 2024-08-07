# Get details of a Citrix QuickCreate AWS Workspaces Directory Connection with directory connection id and account id
data "citrix_quickcreate_aws_workspaces_directory_connection" "example_aws_workspaces_directory_connection" {
    id      = "00000000-0000-0000-0000-000000000000"
    account = "00000000-0000-0000-0000-000000000000"
}

# Get details of a Citrix QuickCreate AWS Workspaces Directory Connection with directory connection name and account id
data "citrix_quickcreate_aws_workspaces_directory_connection" "example_aws_workspaces_directory_connection" {
    name    = "exampleDirectoryConnectionName"
    account = "00000000-0000-0000-0000-000000000000"
}
