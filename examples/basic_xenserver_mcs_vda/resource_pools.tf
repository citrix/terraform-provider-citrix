resource "citrix_xenserver_hypervisor_resource_pool" "example-xenserver-rp" {
    name        = var.resource_pool_name
    hypervisor  = citrix_xenserver_hypervisor.example-xenserver-hypervisor.id
    networks    = var.xenserver_networks
    storage     = [
        {
            storage_name = var.xenserver_storage_name
        }
    ]
    temporary_storage = [
        {
            storage_name = var.xenserver_temporary_storage_name
        }
    ]
    use_local_storage_caching = false
}