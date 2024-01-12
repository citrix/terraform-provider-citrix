resource "citrix_daas_aws_hypervisor_resource_pool" "example-aws-hypervisor-resource-pool" {
    name                = "example-aws-hypervisor-resource-pool"
    hypervisor          = citrix_daas_aws_hypervisor.example-aws-hypervisor.id
    subnets            = [
        "10.0.1.0/24",
    ]
    vpc   = "{VPC name}"
    availability_zone = "us-east-2a"
}