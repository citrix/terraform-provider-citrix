resource "citrix_nutanix_hypervisor_resource_pool" "example-nutanix-hypervisor-resource-pool" {
    name                = "example-nutanix-hypervisor-resource-pool"
    hypervisor          = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id 
    networks    = [
        "<network 1 name>",
        "<network 2 name>"
    ]
}