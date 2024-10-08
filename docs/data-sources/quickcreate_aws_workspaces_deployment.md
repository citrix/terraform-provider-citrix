---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_quickcreate_aws_workspaces_deployment Data Source - citrix"
subcategory: "DaaS Quick Deploy - AWS WorkSpaces Core"
description: |-
  Data source to get details of an AWS WorkSpaces deployment.
---

# citrix_quickcreate_aws_workspaces_deployment (Data Source)

Data source to get details of an AWS WorkSpaces deployment.

## Example Usage

```terraform
# Get details of a Citrix QuickCreate AWS WorkSpaces Deployment using deployment id
data "citrix_quickcreate_aws_workspaces_deployment" "example-deployment" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get details of a Citrix QuickCreate AWS WorkSpaces Deployment using deployment name and account id
data "citrix_quickcreate_aws_workspaces_deployment" "example-deployment" {
  account_id = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account_role_arn.id
  name       = "exampleDeploymentName"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) GUID of the account.
- `id` (String) GUID identifier of the deployment.
- `name` (String) Name of the deployment.

### Read-Only

- `directory_connection_id` (String) GUID of the directory connection.
- `image_id` (String) GUID of the image.