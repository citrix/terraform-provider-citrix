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
hypervisor_name = "example-aws-hyperv"
aws_api_key = "<AWS API key>"
aws_secret_key = "<AWS API secret key>"
aws_region = "us-east-1"

# machine_catalogs.tf variables
machine_catalog_name = "example-aws-catalog"
domain_fqdn = "<DomainFQDN>"
domain_ou = "<DomainOU>"
domain_service_account = "<Admin Username>"
domain_service_account_password = "<Admin Password>"
aws_ami_id = "<AWS AMI ID>"
aws_ami_name =  "<AWS AMI Name>"
aws_service_offering = "t2.small"
machine_catalog_naming_scheme = "ctx-aws-##"

# resource_pools.tf variables
resource_pool_name = "example-aws-resource-pool"
aws_subnets = ["<AWS Subnet Mask>"]
aws_vpc = "<AWS VPC Name>"
aws_availability_zone = "us-east-1a"

# zones.tf variables
zone_name = "example-aws-zone"