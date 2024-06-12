resource "citrix_nutanix_hypervisor_resource_pool" "example-nutanix-rp" {
    name                = var.resource_pool_name
    hypervisor          = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id 
    networks            = var.nutanix_networks
}