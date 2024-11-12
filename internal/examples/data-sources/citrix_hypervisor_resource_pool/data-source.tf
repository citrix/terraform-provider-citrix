# Get Hypervisor Resource Pool of any connection type resource by name and the hypervisor name it belongs to
data "citrix_hypervisor_resource_pool" "azure-resource-pool" {
    name = "azure-rp"
    hypervisor_name = "azure-hyperv"
}

# Get Hypervisor Resource Pool of any connection type resource by id and the hypervisor name it belongs to
data "citrix_hypervisor_resource_pool" "azure-resource-pool" {
    id = "00000000-0000-0000-0000-000000000000"
    hypervisor_name = "azure-hyperv"
}