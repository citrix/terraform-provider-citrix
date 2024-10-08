---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_quickcreate_aws_workspaces_directory_connection Resource - citrix"
subcategory: "DaaS Quick Deploy - AWS WorkSpaces Core"
description: |-
  Manages an AWS WorkSpaces directory connection.
---

# citrix_quickcreate_aws_workspaces_directory_connection (Resource)

Manages an AWS WorkSpaces directory connection.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `account` (String) ID of the account the directory connection is associated with.
- `default_ou` (String) Default OU for VDAs in the directory connection.
- `directory` (String) ID of the AWS directory the directory connection is associated with.
- `name` (String) Name of the directory connection.
- `security_group` (String) ID of the security group the directory connection is associated with.
- `subnets` (Set of String) IDs of the subnets the directory connection is associated with.

### Optional

- `resource_location` (String) ID of the resource location the directory connection is associated with. Only one of `resource_location` and `zone` attributes can be specified.
- `tenancy` (String) Tenancy of the directory connection. Possible values are `SHARED` and `DEDICATED`. Defaults to `DEDICATED`.
- `user_enabled_as_local_administrator` (Boolean) Enable users to be local administrators. Defaults to `false`.
- `zone` (String) ID of the zone the directory connection is associated with. Only one of `zone` and `resource_location` attributes can be specified.

### Read-Only

- `id` (String) GUID identifier of the directory connection.

## Import

Import is supported using the following syntax:

```shell
# Quick Deploy AWS WorkSpaces Directory Connection can be imported by specifying the Account GUID and Directory Connection GUID separated by a comma.
terraform import citrix_quickcreate_aws_workspaces_directory_connection.example_aws_workspaces_directory_connection 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
```