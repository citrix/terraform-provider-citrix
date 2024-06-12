resource "citrix_aws_hypervisor_resource_pool" "example-aws-rp" {
    name                = var.resource_pool_name
    hypervisor          = citrix_aws_hypervisor.example-aws-hypervisor.id
    subnets             = var.aws_subnets
    vpc                 = var.aws_vpc
    availability_zone   = var.aws_availability_zone
}
