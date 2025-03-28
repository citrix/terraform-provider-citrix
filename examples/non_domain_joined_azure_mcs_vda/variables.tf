# citrix.tf variables
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

# For Citrix Cloud customers: API key client id and secret are needed to interact with Citrix DaaS APIs. These can be created/found under Identity and Access Management > API Access
variable provider_client_id {
  description = "The API key client id for Citrix Cloud customer."
  type        = string
  default     = ""
}

variable provider_client_secret {
  description = "The API key client secret for Citrix Cloud customer."
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
  default     = []
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


# images.tf varaiables
variable "image_definition_name" {
  description = "Name of the Image Definition to create"
  type        = string
  default     = "example-image-definition"
}

variable "azure_service_offering" {
  description = "Azure VM service offering SKU"
  type        = string
  default     = "Standard_D2_v2"
}

variable "azure_storage_type" {
  description = "Azure storage type"
  type        = string
  default     = "Standard_LRS"
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


# machine_catalogs.tf variables
variable "machine_catalog_name" {
  description = "Name of the Machine Catalog to create"
  type        = string
  default     = "example-azure-catalog"
}

variable "machine_catalog_naming_scheme" {
  description = "Machine Catalog naming scheme"
  type        = string
  default     = "ctx-non-dj-##"
}


# policy_sets.tf has no variables


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


# resource_location.tf variables
variable "resource_location_name" {
  description = "Name of the Citrix Cloud Resource Location to create."
  type        = string
}
