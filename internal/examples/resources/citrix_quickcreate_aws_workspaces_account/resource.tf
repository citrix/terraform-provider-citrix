# QuickCreate AWS Workspaces Account with AWS Role ARN
resource "citrix_quickcreate_aws_workspaces_account" "example_aws_workspaces_account_role_arn" {
    name                    = "exampe-aws-workspaces-account-role-arn"
    aws_region              = "us-east-1"
    aws_role_arn            = "<AWS Role ARN>"
}

# QuickCreate AWS Workspaces Account with AWS Access Key and Secret Access Key
resource "citrix_quickcreate_aws_workspaces_account" "example_aws_workspaces_account_access_key" {
    name                    = "exampe-aws-workspaces-account-access-key"
    aws_region              = "us-east-1"
    aws_access_key_id       = "<AWS Access Key ID>"
    aws_secret_access_key   = "<AWS Secret Access Key>"
}