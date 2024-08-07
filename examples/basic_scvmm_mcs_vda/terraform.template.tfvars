# citrix.tf variables, uncomment the ones you need for on-premises or cloud
provider_hostname = "<DDC public IP / hostname>" # on-premises only
provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrx Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# delivery_groups.tf variables
delivery_group_name = "example-delivery-group"
allow_list = ["DOMAIN\\user1", "DOMAIN\\user2"]
block_list = ["DOMAIN\\user3", "DOMAIN\\user4"]

# hypervisors.tf variables
hypervisor_name = "example-scvmm-hyperv"
scvmm_username = "<Username>"
scvmm_password = "<Password>"
scvmm_addresses = ["<FQDN for SCVMM>"]

# machine_catalogs.tf variables
machine_catalog_name = "example-scvmm-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
scvmm_master_image = "<Image VM or snapshot name>"
scvmm_cpu_count = 2
scvmm_memory_size = 4096
machine_catalog_naming_scheme = "ctx-scvmm-##"

# resource_pools.tf variables
resource_pool_name = "example-scvmm-resource-pool"
scvmm_networks = ["<Network 1 name>", "<Network 2 name>"]
scvmm_host = "<Host FQDN or IP>"
scvmm_storage_name = "<Local or shared storage name>"
scvmm_temporary_storage_name = "<Local or shared storage name>"

# zones.tf variables
zone_name = "example-scvmm-zone"
# scvmm_resource_location_id = "<SCVMM Resource Location ID>" # cloud only