# Amazon WorkSpaces Core Hypervisor
resource "citrix_amazon_workspaces_core_hypervisor" "example-amazon-workspaces-core-hypervisor-using-api-key" {
    name              = "example-amazon-workspaces-core-hypervisor"
    zone              = "<Zone Id>"
    api_key           = var.aws_account_access_key # AWS account Access Key from variable
    secret_key        = var.aws_account_secret_key # AWS account Secret Key from variable
    region            = "us-east-1"
}

resource "citrix_amazon_workspaces_core_hypervisor" "example-amazon-workspaces-core-hypervisor-using-role-based-auth" {
    name              = "example-amazon-workspaces-core-hypervisor"
    zone              = "<Zone Id>"
    use_iam_role      = true
    region            = "us-east-1"
}