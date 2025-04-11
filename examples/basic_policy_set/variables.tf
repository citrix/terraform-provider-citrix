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


# policy_set_v2.tf variables
variable "scope_id1" {
  description = "First scope id of the policy set"
  type        = string
}

variable "scope_id2" {
  description = "Second scope id of the policy set"
  type        = string
}

variable "delivery_group_id1" {
  description = "ID of the first delivery group for the policy set to apply to"
  type        = string
}

variable "delivery_group_id2" {
  description = "ID of the second delivery group for the policy set to apply to"
  type        = string
}

# policy_filter.tf variables
variable "policy_filter_client_ip" {
  description = "IP of the client for the client ip policy filter"
  type        = string
}

variable "policy_filter_client_name" {
  description = "Name of the client for the client name policy filter"
  type        = string
}

variable "policy_filter_client_platform" {
  description = "Platform of the client for the client platform policy filter"
  type        = string
}

variable "policy_filter_delivery_group_id" {
  description = "ID of the delivery group for the delivery group policy filter"
  type        = string
}

variable "policy_filter_delivery_group_type" {
  description = "Type of the delivery group for the delivery group type policy filter. Values can be `Private`, `PrivateApp`, `Shared`, and `SharedApp`."
  type        = string
}

variable "policy_filter_ou" {
  description = "Organizational unit for the ou policy filter"
  type        = string
}

variable "policy_filter_tag_id" {
  description = "ID of the tag for the tag policy filter"
  type        = string
}

variable "policy_filter_user_sid" {
  description = "SID of the user for the user policy filter"
  type        = string
}

