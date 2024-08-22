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
hypervisor_name = "example-gcp-hyperv"
gcp_service_account_id = "<GCP service account ID>"
gcp_service_account_credentials = "<GCP service account private key>"

# machine_catalogs.tf variables
machine_catalog_name = "example-gcp-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
gcp_storage_type = "pd-standard"
gcp_master_image = "<Image template VM name>"
gcp_availability_zones = "<project name>:<region>:<availability zone1>,<project name>:<region>:<availability zone2>,..."
machine_catalog_naming_scheme = "ctx-gcp-##"

# resource_pools.tf variables
resource_pool_name = "example-gcp-resource-pool"
gcp_project_name = "<Project Name>"
gcp_vpc_region = "<VNet region>"
gcp_vpc = ["<Subnet name>"]
gcp_subnets =  "<VPC name>"

# zones.tf variables
zone_name = "example-gcp-zone"
# gcp_resource_location_id = "<GCP Resource Location ID>" # cloud only