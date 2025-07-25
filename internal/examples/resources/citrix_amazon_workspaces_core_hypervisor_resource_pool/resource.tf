resource "citrix_amazon_workspaces_core_hypervisor_resource_pool" "example-amazon-workspaces-core-hypervisor-resource-pool" {
    name                = "example-amazon-workspaces-core-hypervisor-resource-pool"
    hypervisor          = citrix_amazon_workspaces_core_hypervisor.example-amazon-workspaces-core-hypervisor.id
    subnets            = [
        "10.0.1.0/24",
    ]
    vpc   = "<VPC name>"
    availability_zone = "us-east-2a"
}