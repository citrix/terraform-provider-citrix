resource "citrix_daas_hypervisor_resource_pool" "example-azure-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-azure-hypervisor.id
    region              = "East US"
	virtual_network_resource_group = "{Resource Group Name}"
    virtual_network     = "{VNet name}"
    subnets     			= [
        "subnet 1",
        "subnet 2"
    ]
}

resource "citrix_daas_hypervisor_resource_pool" "example-aws-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-aws-hypervisor.id
    subnets            = [
        "10.0.1.0/24",
    ]
    virtual_network   = "{VPC name}"
    availability_zone = "us-east-2a"
}

resource "citrix_daas_hypervisor_resource_pool" "example-gcp-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-gcp-hypervisor.id
    region             = "us-east1"
    subnets             = [
        "us-east1",
    ]
    virtual_network    = "{VPC name}"
}