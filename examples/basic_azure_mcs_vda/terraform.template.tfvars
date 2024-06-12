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
hypervisor_name = "example-azure-hyperv"
azure_application_id = "<Azure SPN client ID>"
azure_application_secret = "<Azure SPN client secret>"
azure_subscription_id = "<Azure Subscription ID>"
azure_tenant_id = "<Azure Tenant ID>"

# machine_catalogs.tf variables
machine_catalog_name = "example-azure-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
azure_service_offering = "Standard_D2_v2"
azure_storage_type = "Standard_LRS"
#azure_image_subscription = "<Azure Subscription ID for image>"
azure_resource_group = "<Resource Group for VDA image>"
azure_master_image = "<Name for VDA image>"
#azure_gallery_name = "<Azure Gallery Name>"
#azure_gallery_image_definition = "<Azure Gallery image definition>"
#azure_gallery_image_version = "<Azure Gallery image version>"
machine_catalog_naming_scheme = "ctx-azure-##"

# resource_pools.tf variables
resource_pool_name = "example-azure-resource-pool"
azure_region = "East US"
azure_vnet_resource_group = "<VNet resource group name>"
azure_vnet = "<VNet name>"
azure_subnets = ["<Subnet name>"]

# zones.tf variables
zone_name = "example-azure-zone"