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
  default     = "example-scvmm-hyperv"
}

variable "scvmm_username" {
  description = "Username to the SCVMM hypervisor"
  type        = string
}

variable "scvmm_password" {
  description = "Password to the SCVMM hypervisor"
  type        = string
  sensitive   = true
}

variable "scvmm_password_format" {
  description = "SCVMM password format"
  type        = string
  default     = "PlainText"
}

variable "scvmm_addresses" {
  description = "List of addresses to the SCVMM hypervisor in the format of \"<FQDN for SCVMM>\""
  type        = list(string)
}


# machine_catalogs.tf variables
variable "machine_catalog_name" {
  description = "Name of the Machine Catalog to create"
  type        = string
  default     = "example-scvmm-catalog"
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

variable "scvmm_master_image" {
  description = "Name of the VM to be used as a master image"
  type        = string
}

variable "scvmm_cpu_count" {
  description = "Number of CPUs per VM created"
  type        = number
  default     = 2
}

variable "scvmm_memory_size" {
  description = "Amount of memory in MB per VM created"
  type        = number
  default     = 4096
}

variable "machine_catalog_naming_scheme" {
  description = "Machine Catalog naming scheme"
  type        = string
  default     = "ctx-scvmm-##"
}


# resource_pools.tf variables
variable "resource_pool_name" {
  description = "Name of the Resource Pool to create"
  type        = string
  default     = "example-scvmm-rp"
}

variable "scvmm_networks" {
  description = "List of network names for the Resource Pool to use"
  type        = list(string)
}

variable "scvmm_host" {
  description = "The name of the host"
  type        = string
}

variable "scvmm_storage_name" {
  description = "Name of the storage"
  type        = string
}

variable "scvmm_temporary_storage_name" {
  description = "Name of the temporary storage"
  type        = string
}

# zones.tf variables
variable "zone_name" {
  description = "Name of the Zone to create. For Citrix Cloud customers the zone should already exist."
  type        = string
}
