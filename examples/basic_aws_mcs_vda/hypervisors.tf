# AWS Hypervisor
resource "citrix_aws_hypervisor" "example-aws-hypervisor" {
    name              = var.hypervisor_name
    zone              = citrix_zone.example-zone.id
    api_key           = var.aws_api_key
    secret_key        = var.aws_secret_key
    region            = var.aws_region
}
