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
  default     = "example-azure-hyperv"
}

variable "azure_application_id" {
  description = "Azure SPN client ID"
  type        = string
}

variable "azure_application_secret" {
  description = "Azure SPN client secret"
  type        = string
  sensitive   = true
}

variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "azure_tenant_id" {
  description = "Azure tenant ID"
  type        = string
}


# machine_catalogs.tf variables
variable "machine_catalog_name" {
  description = "Name of the Machine Catalog to create"
  type        = string
  default     = "example-azure-catalog"
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

variable "azure_service_offering" {
  description = "Azure VM service offering SKU"
  type        = string
  default     = "Standard_D2_v2"
}

# variable "azure_image_subscription" {
#   description = "Azure subscription ID for the image, not needed if image is in the same subscription as the hypervisor"
#   type        = string
# }

# For Azure master image from managed disk or snapshot
variable "azure_resource_group" {
  description = "Azure resource group containing the master image"
  type        = string
}

variable "azure_master_image" {
  description = "Name of the master image managed disk or snapshot"
  type        = string
}

# For Azure image gallery
# variable "azure_gallery_name" {
#   description = "Azure gallery image name"
#   type        = string
# }

# variable "azure_gallery_image_definition" {
#   description = "Azure gallery image definition"
#   type        = string
# }

# variable "azure_gallery_image_version" {
#   description = "Azure gallery image version"
#   type        = string
#   default     = "1.0.0"
# }

variable "azure_storage_type" {
  description = "Azure storage type"
  type        = string
  default     = "Standard_LRS"
}

variable "machine_catalog_naming_scheme" {
  description = "Machine Catalog naming scheme"
  type        = string
  default     = "ctx-azure-##"
}


# resource_pools.tf variables
variable "resource_pool_name" {
  description = "Name of the Resource Pool to create"
  type        = string
  default     = "example-azure-rp"
}

variable "azure_region" {
  description = "Azure region for the Resource Pool"
  type        = string
  default     = "East US"
}

variable "azure_vnet_resource_group" {
  description = "Name of the Azure virtual network resource group"
  type        = string
}

variable "azure_vnet" {
  description = "Name of the Azure virtual network"
  type        = string
}

variable "azure_subnets" {
  description = "List of Azure subnets"
  type        = list(string)
}


# zones.tf variables
## CVAD zone configuration
variable "zone_name" {
  description = "Name of the Zone to create. For Citrix Cloud customers please use aws_resource_location_id."
  type        = string
  default     = ""
}

## DaaS zone configuration
variable "azure_resource_location_id" {
  description = "The resource location id of the Azure resource location for creating zone, which should have connector(s) provisioned."
  type        = string
  default     = ""
}
