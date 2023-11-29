resource "citrix_daas_hypervisor_resource_pool" "example-azure-rp" {
    name                = "example-azure-rp"
    hypervisor          = citrix_daas_hypervisor.example-azure-hypervisor.id
    region              = "<VNet region>"
    virtual_network_resource_group = "<VNet resource group name>"
    virtual_network     = "<VNet name>"
    subnets              = [
        "<Subnet name>"
    ]
}