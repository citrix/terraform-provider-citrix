# citrix.tf variables, uncomment the ones you need for on-premises or cloud
provider_hostname = "<DDC public IP / hostname>" # on-premises only
provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrix Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# delivery_groups.tf variables
delivery_group_name = "example-delivery-group"
allow_list = ["DOMAIN\\user1", "DOMAIN\\user2"]
block_list = ["DOMAIN\\user3", "DOMAIN\\user4"]

# hypervisors.tf variables
hypervisor_name = "example-vsphere-hyperv"
vsphere_username = "<Username>"
vsphere_password = "<Password>"
vsphere_addresses = ["http://<IP address or hostname for vSphere>"]

# machine_catalogs.tf variables
machine_catalog_name = "example-vsphere-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
vsphere_master_image_vm = "<Image VM or snapshot name>"
vsphere_cpu_count = 2
vsphere_memory_size = 4096
machine_catalog_naming_scheme = "ctx-vsphere-##"

# resource_pools.tf variables
resource_pool_name = "example-vsphere-resource-pool"
vsphere_networks = ["<Network 1 name>", "<Network 2 name>"]
vsphere_cluster_datacenter = "<Datacenter name>"
vsphere_cluster_name = "<Cluster name>"
vsphere_host = "<Host FQDN or IP>"
vsphere_storage_name = "<Local or shared storage name>"
vsphere_temporary_storage_name = "<Local or shared storage name>"

# zones.tf variables
zone_name = "example-vsphere-zone"
# vsphere_resource_location_id = "<vSphere Resource Location ID>" # cloud only