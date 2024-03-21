resource "citrix_vsphere_hypervisor_resource_pool" "example-vsphere-rp" {
    name                = "example-vsphere-rp"
    hypervisor          = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
    cluster             = {
        datacenter = "<datacenter_name>"
        cluster_name = "<cluster_name>"
        host = "<host_fqdn_or_ip>"
    }
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