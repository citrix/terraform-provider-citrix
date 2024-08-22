# citrix.tf variables
provider_customer_id   = "<Citrix Cloud CustomerID>"
provider_client_id     = "<Citrix Cloud secure client ID>"
provider_client_secret = "<Citrix Cloud secure client secret>"

# resource_location.tf variables
resource_locaiton_name = "example-resource-location"

# account.tf variables
account_name     = "example-aws-workspaces-account"
account_region   = "us-east-1"
account_role_arn = "arn:aws:iam::123456789012:role/role-name"

# image.tf variables
image_name        = "example-aws-workspaces-image"
image_id          = "ami-1234567890abcdef0"
image_description = "Example AWS WorkSpaces image"

# directory_connection.tf variables
directory_connection_name              = "example-aws-workspaces-directory-connection"
directory_connection_id                = "d-1234567890abcdef0"
directory_connection_subnet_1          = "subnet-1234567890abcdef0"
directory_connection_subnet_2          = "subnet-1234567890abcdef1"
directory_connection_security_group_id = "sg-1234567890abcdef0"
directory_connection_ou                = "OU=VDAs,OU=Computers,OU=example,DC=example,DC=local"

# deployment.tf variables
deployment_name       = "example-aws-workspaces-deployment"
deployment_usernames  = ["user0001", "user0002"]
