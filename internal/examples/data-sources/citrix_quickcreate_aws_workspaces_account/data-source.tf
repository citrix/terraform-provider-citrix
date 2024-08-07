# Get details of a Citrix QuickCreate AWS Workspaces Account using the account GUID
data "citrix_quickcreate_aws_workspaces_account" "example-account" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get details of a Citrix QuickCreate AWS Workspaces Account using the account name
data "citrix_quickcreate_aws_workspaces_account" "example-account" {
  name = "exampleAccountName"
}