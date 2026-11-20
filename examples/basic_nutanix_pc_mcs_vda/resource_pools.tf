resource "citrix_nutanix_pc_hypervisor_resource_pool" "example-nutanix-pc-rp" {
    name       = var.resource_pool_name
    hypervisor = citrix_nutanix_pc_hypervisor.example-nutanix-pc-hypervisor.id
}
