resource "citrix_daas_hypervisor_resource_pool" "example-azure-hypervisor-resource_pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-azure-hypervisor.id
    region              = "East US"
	virtual_network_resource_group = "{Resource Group Name}"
    virtual_network     = "{VNet name}"
    subnet     			= [
        "subnet 1",
        "subnet 2"
    ]
}

resource "citrix_daas_hypervisor_resource_pool" "example-aws-hypervisor-resource_pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-aws-hypervisor.id
    subnet            = [
        "10.0.1.0/24",
    ]
    virtual_network   = "{VPC name}"
    availability_zone = "us-east-2a"
}

resource "citrix_daas_hypervisor_resource_pool" "example-gcp-hypervisor-resource_pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-gcp-hypervisor.id
    region             = "us-east1"
    subnet             = [
        "us-east1",
    ]
    virtual_network    = "{VPC name}"
}