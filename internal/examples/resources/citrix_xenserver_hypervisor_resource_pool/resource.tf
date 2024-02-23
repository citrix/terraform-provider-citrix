resource "citrix_xenserver_hypervisor_resource_pool" "example-xenserver-hypervisor-resource-pool" {
    name                = "example-xenserver-hypervisor-resource-pool"
    hypervisor          = citrix_xenserver_hypervisor.example-xenserver-hypervisor.id
    networks    = [
        "<network 1 name>",
        "<network 2 name>"
    ]
    storage     = [
        "<local or shared storage name>"
    ]
    temporary_storage = [
        "<local or shared storage name>"
    ]
    use_local_storage_caching = false
}