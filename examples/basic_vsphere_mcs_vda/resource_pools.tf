resource "citrix_vsphere_hypervisor_resource_pool" "example-vsphere-rp" {
    name                = var.resource_pool_name
    hypervisor          = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
    cluster             = {
        datacenter = var.vsphere_cluster_datacenter
        cluster_name = var.vsphere_cluster_name
        host = var.vsphere_cluster_host
    }
    networks    = var.vsphere_networks
    storage     = [
        {
            storage_name = var.vsphere_storage_name
        }
    ]
    temporary_storage = [
        {
            storage_name = var.vsphere_temporary_storage_name
        }
    ]
    use_local_storage_caching = false
}