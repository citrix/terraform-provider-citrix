resource "citrix_vsphere_hypervisor_resource_pool" "example-vsphere-hypervisor-resource-pool" {
    name                = "example-vsphere-hypervisor-resource-pool"
    hypervisor          = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
    cluster             = {
        datacenter = "<datacenter_name>"
        cluster_name = "<cluster_name>"
        # host = "<host_fqdn_or_ip>" // Use one of host or cluster
    }
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