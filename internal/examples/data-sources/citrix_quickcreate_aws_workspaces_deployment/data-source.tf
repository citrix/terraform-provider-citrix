# Get details of a Citrix QuickCreate AWS WorkSpaces Deployment using deployment id
data "citrix_quickcreate_aws_workspaces_deployment" "example-deployment" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get details of a Citrix QuickCreate AWS WorkSpaces Deployment using deployment name and account id
data "citrix_quickcreate_aws_workspaces_deployment" "example-deployment" {
  account_id = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account_role_arn.id
  name       = "exampleDeploymentName"
}
