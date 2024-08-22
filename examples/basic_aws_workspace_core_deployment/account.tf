resource "citrix_quickcreate_aws_workspaces_account" "example_aws_workspaces_account" {
    name                    = var.account_name
    aws_region              = var.account_region
    aws_role_arn            = var.account_role_arn
}
