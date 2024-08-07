# AWS Hypervisor
resource "citrix_aws_hypervisor" "example-aws-hypervisor" {
    name              = "example-aws-hypervisor"
    zone              = "<Zone Id>"
    api_key           = var.aws_account_access_key # AWS account Access Key from variable
    secret_key        = var.aws_account_secret_key # AWS account Secret Key from variable
    region            = "us-east-2"
}