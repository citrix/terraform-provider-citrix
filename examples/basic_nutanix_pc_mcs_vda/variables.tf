# citrix.tf variables
## On-Premises customer provider settings
variable provider_hostname {
  description = "The hostname of the Citrix Virtual Apps and Desktops Delivery Controller."
  type        = string
  default     = "" # Leave this variable empty for Citrix Cloud customer.
}

variable provider_domain_fqdn {
  description = "The domain FQDN of the on-premises Active Directory."
  type        = string
  default     = null # Leave this variable empty for Citrix Cloud customer.
}

variable provider_disable_ssl_verification {
  description = "Disable SSL verification for the Citrix Virtual Apps and Desktops Delivery Controller."
  type        = bool
  default     = false # Set this field to true if DDC does not have a valid SSL certificate configured. Omit this variable for Citrix Cloud customer. 
}

## Citrix Cloud customer provider settings
variable provider_customer_id {
  description = "The customer id of the Citrix Cloud customer."
  type        = string
  default     = "" # Set your customer id for Citrix Cloud customer. Omit this variable for On-Premises customer.
}

variable provider_environment {
  description = "The environment of the Citrix Cloud customer."
  type        = string
  default     = "Production" # Use "Japan" for Citrix Cloud customers in Japan region. Omit this variable for On-Premises customer.
}

# Common provider settings
# For On-Premises customers: Domain Admin username and password are needed to interact with the Citrix Virtual Apps and Desktops Delivery Controller.
# For Citrix Cloud customers: API key client id and secret are needed to interact with Citrix DaaS APIs. These can be created/found under Identity and Access Management > API Access
variable provider_client_id {
  description = "The Domain Admin username of the on-premises Active Directory / The API key client id for Citrix Cloud customer."
  type        = string
  default     = ""
}

variable provider_client_secret {
  description = "The Domain Admin password of the on-premises Active Directory / The API key client secret for Citrix Cloud customer."
  type        = string
  default     = ""
}


# delivery_groups.tf variables
variable "delivery_group_name" {
  description = "Name of the Delivery Group to create"
  type        = string
  default     = "example-delivery-group"
}

variable "allow_list" {
  description = "List of users to allow for the Delivery Group in DOMAIN\\username format"
  type        = list(string)
}


# hypervisors.tf variables
variable "hypervisor_name" {
  description = "Name of the Hypervisor to create"
  type        = string
  default     = "example-nutanix-pc-hyperv"
}

variable "nutanix_pc_username" {
  description = "Username to the Nutanix Prism Central"
  type        = string
}

variable "nutanix_pc_password" {
  description = "Password to the Nutanix Prism Central"
  type        = string
  sensitive   = true
}

variable "nutanix_pc_password_format" {
  description = "Nutanix Prism Central password format"
  type        = string
  default     = "PlainText"
}

variable "nutanix_pc_addresses" {
  description = "List of Prism Central addresses: host name or IP address, optionally followed by :port (default 9440), without http:// or https://"
  type        = list(string)
}

variable "nutanix_pc_ssl_thumbprints" {
  description = "SSL certificate thumbprints of Prism Central, required when its certificate isn't trusted by the Cloud Connector or Delivery Controller"
  type        = list(string)
  default     = null
}


# machine_catalogs.tf variables
variable "machine_catalog_name" {
  description = "Name of the Machine Catalog to create"
  type        = string
  default     = "example-nutanix-pc-catalog"
}

variable "domain_fqdn" {
  description = "Domain FQDN"
  type        = string
}

variable "domain_ou" {
  description = "Domain organizational unit"
  type        = string
  default     = null
}

variable "domain_service_account" {
  description = "Domain service account with permissions to create machine accounts"
  type        = string
}

variable "domain_service_account_password" {
  description = "Domain service account password"
  type        = string
  sensitive   = true
}

variable "nutanix_pc_cluster" {
  description = "Name of the Prism Central-managed cluster to provision machines in"
  type        = string
}

variable "nutanix_pc_master_image" {
  description = "Name of the Prism Central template used as the master image"
  type        = string
}

variable "nutanix_pc_master_image_version" {
  description = "Name of the template version used as the master image"
  type        = string
}

variable "nutanix_pc_network" {
  description = "Name of the subnet in the selected cluster to attach to the first NIC"
  type        = string
}

variable "nutanix_pc_cpu_count" {
  description = "Number of CPU sockets per VM created"
  type        = number
  default     = 2
}

variable "nutanix_pc_cores_per_cpu_count" {
  description = "Number of cores per CPU socket per VM created"
  type        = number
  default     = 2
}

variable "nutanix_pc_memory_size" {
  description = "Amount of memory in MB per VM created"
  type        = number
  default     = 4096
}

variable "machine_catalog_naming_scheme" {
  description = "Machine Catalog naming scheme"
  type        = string
  default     = "ctx-ntxpc-##"
}


# resource_pools.tf variables
variable "resource_pool_name" {
  description = "Name of the Resource Pool to create"
  type        = string
  default     = "example-nutanix-pc-rp"
}


# zones.tf variables
## CVAD zone configuration
variable "zone_name" {
  description = "Name of the Zone to create. For Citrix Cloud customers please use nutanix_resource_location_id."
  type        = string
  default     = ""
}

## DaaS zone configuration
variable "nutanix_resource_location_id" {
  description = "The resource location id of the Nutanix resource location for creating zone, which should have connector(s) provisioned."
  type        = string
  default     = ""
}
