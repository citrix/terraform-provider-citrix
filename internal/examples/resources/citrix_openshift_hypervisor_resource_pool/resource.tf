resource "citrix_openshift_hypervisor_resource_pool" "example-openshift-hypervisor-resource-pool" {
    name                = "example-openshift-hypervisor-resource-pool"
    hypervisor          = citrix_openshift_hypervisor.example-openshift-hypervisor.id
    namespace           = "<namespace_name>"
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
}