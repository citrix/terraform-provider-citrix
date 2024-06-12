resource "citrix_azure_hypervisor_resource_pool" "example-azure-rp" {
    name                = var.resource_pool_name
    hypervisor          = citrix_azure_hypervisor.example-azure-hypervisor.id
    region              = var.azure_region
    virtual_network_resource_group = var.azure_vnet_resource_group
    virtual_network     = var.azure_vnet
    subnets             = var.azure_subnets
}