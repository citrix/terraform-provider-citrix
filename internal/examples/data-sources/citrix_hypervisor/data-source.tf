# Get Hypervisor resource of any connection type by name
data "citrix_hypervisor" "azure-hypervisor" {
    name = "azure-hyperv"
}

# Get Hypervisor resource of any connection type by id
data "citrix_hypervisor" "azure-hypervisor" {
    id = "00000000-0000-0000-0000-000000000000"
}