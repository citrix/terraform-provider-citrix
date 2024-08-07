# citrix.tf variables, uncomment the ones you need for on-premises or cloud
provider_hostname = "<DDC public IP / hostname>" # on-premises only
provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrx Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# delivery_groups.tf variables
delivery_group_name = "example-delivery-group"
allow_list = ["DOMAIN\\user1", "DOMAIN\\user2"]

# hypervisors.tf variables
hypervisor_name = "example-xen-hyperv"
xenserver_username = "<Username>"
xenserver_password = "<Password>"
xenserver_addresses = ["http://<IP address or hostname for XenServer>"]

# machine_catalogs.tf variables
machine_catalog_name = "example-xen-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
xenserver_master_image_vm = "<Image VM or snapshot name>"
xenserver_cpu_count = 2
xenserver_memory_size = 4096
machine_catalog_naming_scheme = "ctx-xen-##"

# resource_pools.tf variables
resource_pool_name = "example-xen-resource-pool"
xenserver_networks = ["<Network 1 name>", "<Network 2 name>"]
xenserver_storage_name = "<Local or shared storage name>"
xenserver_temporary_storage_name = "<Local or shared storage name>"

# zones.tf variables
zone_name = "example-xen-zone"
# xenserver_resource_location_id = "<XenServer Resource Location ID>" # cloud only