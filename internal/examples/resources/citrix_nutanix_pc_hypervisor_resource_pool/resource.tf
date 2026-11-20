resource "citrix_nutanix_pc_hypervisor_resource_pool" "example-nutanix-pc-hypervisor-resource-pool" {
    name       = "example-nutanix-pc-hypervisor-resource-pool"
    hypervisor = citrix_nutanix_pc_hypervisor.example-nutanix-pc-hypervisor.id
}
