resource "citrix_aws_hypervisor_resource_pool" "example-aws-rp" {
    name                = "example-aws-rp"
    hypervisor          = citrix_aws_hypervisor.example-aws-hypervisor.id
    subnets             = [
        "<AWS Subnet Mask>",
    ]
    vpc   = "<AWS VPC Name>"
    availability_zone = "us-east-1a"
}
