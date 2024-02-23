# AWS Hypervisor
resource "citrix_aws_hypervisor" "example-aws-hypervisor" {
    name              = "example-aws-hypervisor"
    zone              = "<Zone Id>"
    api_key           = "<AWS account Access Key>"
    secret_key        = "<AWS account Secret Key>"
    region            = "us-east-2"
}