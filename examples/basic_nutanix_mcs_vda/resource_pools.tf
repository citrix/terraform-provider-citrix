resource "citrix_nutanix_hypervisor_resource_pool" "example-nutanix-rp" {
    name                = "example-nutanix-rp"
    hypervisor          = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id 
    networks            = [
        "<network 1 name>",
        "<network 2 name>"
    ]
}