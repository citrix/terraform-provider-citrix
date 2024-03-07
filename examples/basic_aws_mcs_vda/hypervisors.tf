# AWS Hypervisor
resource "citrix_aws_hypervisor" "example-aws-hypervisor" {
    name              = "example-aws-hyperv"
    zone              = citrix_zone.example-zone.id
    api_key           = "{AWS API access key}"
    secret_key        = "{AWS API secret key}"
    region            = "{AWS region}"
}
