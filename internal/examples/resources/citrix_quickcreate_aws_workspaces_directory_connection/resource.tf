resource "citrix_quickcreate_aws_workspaces_directory_connection" "test_aws_directory_connection_with_resource_location" {
    name                                = "example-directory-connection"
    account                             = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    resource_location                   = citrix_cloud_resource_location.example-resource-location.id
    directory                           = "d-012345abcd"
    subnets                             = ["subnet-00000000000000000", "subnet-11111111111111111"]
    tenancy                             = "DEDICATED"
    security_group                      = "sg-00000000000000000"
    default_ou                          = "OU=VDAs,OU=Computers,OU=test-out,DC=test,DC=local"
    user_enabled_as_local_administrator = false
}

resource "citrix_quickcreate_aws_workspaces_directory_connection" "test_aws_directory_connection_with_resource_location_with_zone" {
    name                                = "example-directory-connection"
    account                             = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    zone                                = citrix_zone.example-cloud-zone.id
    directory                           = "d-012345abcd"
    subnets                             = ["subnet-00000000000000000", "subnet-11111111111111111"]
    tenancy                             = "DEDICATED"
    security_group                      = "sg-00000000000000000"
    default_ou                          = "OU=VDAs,OU=Computers,OU=test-out,DC=test,DC=local"
    user_enabled_as_local_administrator = false
}
