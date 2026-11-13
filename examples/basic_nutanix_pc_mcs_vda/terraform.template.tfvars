# citrix.tf variables, uncomment the ones you need for on-premises or cloud
provider_hostname = "<DDC public IP / hostname>" # on-premises only
provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrix Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# delivery_groups.tf variables
delivery_group_name = "example-delivery-group"
allow_list = ["DOMAIN\\user1", "DOMAIN\\user2"]

# hypervisors.tf variables
hypervisor_name = "example-nutanix-pc-hyperv"
nutanix_pc_username = "<Username>"
nutanix_pc_password = "<Password>"
nutanix_pc_addresses = ["<IP address or hostname of Prism Central>"]
# nutanix_pc_ssl_thumbprints = ["<SSL thumbprint of Prism Central>"]

# machine_catalogs.tf variables
machine_catalog_name = "example-nutanix-pc-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
nutanix_pc_cluster = "<Prism Central cluster name>"
nutanix_pc_master_image = "<Template name>"
nutanix_pc_master_image_version = "<Template version name>"
nutanix_pc_network = "<Subnet name in the cluster>"
nutanix_pc_cpu_count = 2
nutanix_pc_cores_per_cpu_count = 2
nutanix_pc_memory_size = 4096
machine_catalog_naming_scheme = "ctx-ntxpc-##"

# resource_pools.tf variables
resource_pool_name = "example-nutanix-pc-resource-pool"

# zones.tf variables
zone_name = "example-nutanix-pc-zone"
# nutanix_resource_location_id = "<Nutanix Resource Location ID>" # cloud only
