# citrix.tf variables
## Citrix Cloud customer provider settings
variable "provider_customer_id" {
  description = "The customer id of the Citrix Cloud customer."
  type        = string
}

variable "provider_environment" {
  description = "The environment of the Citrix Cloud customer."
  type        = string
  default     = "Production" # Use "Japan" for Citrix Cloud customers in Japan region.
}

## Common provider settings
## API key client id and secret are needed to interact with Citrix DaaS APIs. These can be created/found under Identity and Access Management > API Access
variable "provider_client_id" {
  description = "The API key client id for Citrix Cloud customer."
  type        = string
}

variable "provider_client_secret" {
  description = "The API key client secret for Citrix Cloud customer."
  type        = string
  sensitive   = true
}

# resource_location.tf variables
variable "resource_location_name" {
  description = "Name of the Citrix Cloud Resource Location to create."
  type        = string
}


# account.tf variables
variable "account_name" {
  description = "The name of the AWS WorkSpaces account."
  type        = string
}

variable "account_region" {
  description = "The region of the AWS WorkSpaces account."
  type        = string
}

variable "account_role_arn" {
  description = "The assume role ARN of the AWS WorkSpaces account."
  type        = string
  sensitive   = true
}

# image.tf variables
variable "image_name" {
  description = "The name of the AWS WorkSpaces image."
  type        = string
}

variable "image_id" {
  description = "The ID of the AWS WorkSpaces image in AWS."
  type        = string
}

variable "image_description" {
  description = "The description of the AWS WorkSpaces image."
  type        = string
}

variable "image_session_support" {
  description = "The session support of the AWS WorkSpaces image."
  type        = string
  default     = "SingleSession"
}

variable "image_operating_system" {
  description = "The operating system of the AWS WorkSpaces image."
  type        = string
  default     = "WINDOWS"
}

variable "image_ingestion_process" {
  description = "The ingestion process of the AWS WorkSpaces image."
  type        = string
  default     = "BYOL_REGULAR_BYOP"
}

# directory_connection.tf variables
variable "directory_connection_name" {
  description = "The name of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_id" {
  description = "The id of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_subnet_1" {
  description = "The first subnet of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_subnet_2" {
  description = "The second subnet of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_tenancy" {
  description = "The tenancy of the AWS WorkSpaces directory connection."
  type        = string
  default     = "DEDICATED"
}

variable "directory_connection_security_group_id" {
  description = "The ID of the security group of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_ou" {
  description = "The OU of the AWS WorkSpaces directory connection."
  type        = string
}

variable "directory_connection_user_as_local_admin" {
  description = "Indicates whether enable users as local administrator in the AWS WorkSpaces directory connection."
  type        = bool
  default     = false
}

# deployment.tf variables
variable "deployment_name" {
  description = "The name of the AWS WorkSpaces deployment."
  type        = string
}

variable "deployment_usernames" {
  description = "A list of users for the AWS WorkSpaces deployment."
  type        = list(string)
}
