# citrix.tf variables, uncomment the ones you need for on-premises or cloud
# provider_hostname = "<DDC public IP / hostname>" # on-premises only
# provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrx Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# delivery_groups.tf variables
delivery_group_name = "example-delivery-group"
allow_list = ["DOMAIN\\user1", "DOMAIN\\user2"]
block_list = ["DOMAIN\\user3", "DOMAIN\\user4"]

# hypervisors.tf variables
hypervisor_name = "example-nutanix-hyperv"
nutanix_username = "<Username>"
nutanix_password = "<Password>"
nutanix_addresses = ["http://<IP address or hostname for Nutanix>"]

# machine_catalogs.tf variables
machine_catalog_name = "example-nutanix-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
nutanix_container = "<Container name>"
nutanix_master_image = "<Image name>"
nutanix_cpu_count = 2
nutanix_core_per_cpu_count = 2
nutanix_memory_size = 4096
machine_catalog_naming_scheme = "ctx-nutanix-##"

# resource_pools.tf variables
resource_pool_name = "example-nutanix-resource-pool"
nutanix_networks = ["<Network 1 name>", "<Network 2 name>"]

# zones.tf variables
zone_name = "example-nutanix-zone"