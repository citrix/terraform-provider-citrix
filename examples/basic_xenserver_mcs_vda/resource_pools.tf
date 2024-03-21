resource "citrix_xenserver_hypervisor_resource_pool" "example-xenserver-rp" {
    name                = "example-xenserver-rp"
    hypervisor          = citrix_xenserver_hypervisor.example-xenserver-hypervisor.id
    networks    = [
        "<network 1 name>",
        "<network 2 name>"
    ]
    storage     = [
        {
            storage_name = "<local or shared storage name>"
        }
    ]
    temporary_storage = [
        {
            storage_name = "<local or shared storage name>"
        }
    ]
    use_local_storage_caching = false
}