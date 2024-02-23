resource "citrix_azure_hypervisor_resource_pool" "example-azure-hypervisor-resource-pool" {
    name                = "example-azure-hypervisor-resource-pool"
    hypervisor          = citrix_azure_hypervisor.example-azure-hypervisor.id
    region              = "East US"
	virtual_network_resource_group = "<Resource Group Name>"
    virtual_network     = "<VNet name>"
    subnets     			= [
        "subnet 1",
        "subnet 2"
    ]
}