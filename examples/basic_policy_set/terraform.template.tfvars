# citrix.tf variables, uncomment the ones you need for on-premises or cloud
provider_hostname = "<DDC public IP / hostname>" # on-premises only
provider_domain_fqdn = "<DomainFqdn>" # on-premises only
provider_client_id = "<Admin Username>" # or Citrix Cloud secure client ID for cloud
provider_client_secret = "<Admin Password>" # or Citrix Cloud secure client secret for cloud
# provider_customer_id = "<Citrix Cloud CustomerID>" # cloud only

# policy_set_v2.tf variables
scope_id1 = "09dd6c92-51d8-42d0-a93b-838490e8e35d" // example value
scope_id2 = "16d65fa7-e67c-4bea-8dfc-e8b4d089c138" // example value
delivery_group_id1 = "275dca7f-38a5-4f61-b124-f35358fe2833" // example value
delivery_group_id2 = "6f4d47d1-16eb-4b57-b744-74f07fa4884d" // example value

# policy_filter.tf variables
policy_filter_client_ip = "192.168.0.1"
policy_filter_client_name = "example-client-name"
policy_filter_client_platform = "Windows"
policy_filter_delivery_group_id = "a298d13d-9197-44bc-8ec9-75e0e4ae8fd8"
policy_filter_delivery_group_type ="Private"
policy_filter_ou = "OU=example,DC=example,DC=local"
policy_filter_tag_id = "8668672a-31ab-49f8-8a42-4babbef18a3c" // example value
policy_filter_user_sid = "S-1-5-21-3623811015-3361044348-30300820-1013" // example value
