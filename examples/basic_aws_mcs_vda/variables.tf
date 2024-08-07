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
  default     = []
}


# hypervisors.tf variables
variable "hypervisor_name" {
  description = "Name of the Hypervisor to create"
  type        = string
  default     = "example-aws-hyperv"
}

variable "aws_api_key" {
  description = "AWS API Key"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS Secret Key"
  type        = string
  sensitive   = true
}

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "us-east-1"
}


# machine_catalogs.tf variables
variable "machine_catalog_name" {
  description = "Name of the Machine Catalog to create"
  type        = string
  default     = "example-aws-catalog"
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

variable "aws_ami_id" {
  description = "AWS AMI ID"
  type        = string
}

variable "aws_ami_name" {
  description = "AWS AMI Name"
  type        = string
}

variable "aws_service_offering" {
  description = "AWS Service Offering"
  type        = string
  default     = "t2.small"
}

variable "machine_catalog_naming_scheme" {
  description = "Machine Catalog naming scheme"
  type        = string
  default     = "ctx-aws-##"
}


# resource_pools.tf variables
variable "resource_pool_name" {
  description = "Name of the Resource Pool to create"
  type        = string
  default     = "example-aws-rp"
}

variable "aws_subnets" {
  description = "List of AWS subnets"
  type        = list(string)
}

variable "aws_vpc" {
  description = "AWS VPC name"
  type        = string
}

variable "aws_availability_zone" {
  description = "AWS Availability Zone"
  type        = string
  default     = "us-east-1a"
}


# zones.tf variables
## CVAD zone configuration
variable "zone_name" {
  description = "Name of the Zone to create. For Citrix Cloud customers please use aws_resource_location_id."
  type        = string
  default     = ""
}

## DaaS zone configuration
variable "aws_resource_location_id" {
  description = "The resource location id of the AWS resource location for creating zone, which should have connector(s) provisioned."
  type        = string
  default     = ""
}
