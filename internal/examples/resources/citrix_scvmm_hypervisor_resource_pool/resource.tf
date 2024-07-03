resource "citrix_scvmm_hypervisor_resource_pool" "example-scvmm-hypervisor-resource-pool" {
    name                = "example-scvmm-hypervisor-resource-pool"
    hypervisor          = citrix_scvmm_hypervisor.example-scvmm-hypervisor.id
    host = "<host name>"
    networks    = [
        "<network 1 name>",
        "<network 2 name>"
    ]
    storage     = [
        {
            storage_name = "<local or shared storage name>"
            superseded = false # Only to be used for updates
        }
    ]
    temporary_storage = [
        {
            storage_name = "<local or shared storage name>"
            superseded = false # Only to be used for updates
        }
    ]
    use_local_storage_caching = false
}