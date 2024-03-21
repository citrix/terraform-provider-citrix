resource "citrix_azure_hypervisor_resource_pool" "example-azure-rp" {
    name                = "example-azure-rp"
    hypervisor          = citrix_azure_hypervisor.example-azure-hypervisor.id
    region              = "East US"
    virtual_network_resource_group = "<VNet resource group name>"
    virtual_network     = "<VNet name>"
    subnets              = [
        "<Subnet name>"
    ]
}