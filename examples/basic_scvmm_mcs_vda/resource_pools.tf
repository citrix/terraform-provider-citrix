resource "citrix_scvmm_hypervisor_resource_pool" "example-scvmm-hypervisor-resource-pool" {
    name                = var.resource_pool_name
    hypervisor          = citrix_scvmm_hypervisor.example-scvmm-hypervisor.id
    host = var.scvmm_host_name
    networks    = var.scvmm_networks
    storage     = [
        {
            storage_name = var.scvmm_storage_name
        }
    ]
    temporary_storage = [
        {
            storage_name = var.scvmm_temporary_storage_name
        }
    ]
    use_local_storage_caching = false
}