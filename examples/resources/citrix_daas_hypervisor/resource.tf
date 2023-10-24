# Azure Hypervisor
resource "citrix_daas_hypervisor" "example-azure-hypervisor" {
    name                = "example-azure-hypervisor"
    connection_type     = "AzureRM"
    zone                = "{Zone Id}"
    active_directory_id = "{Azure Tenant Id}"
    subscription_id     = "{Azure Subscription Id}"
    application_secret  = "{Azure Client Secret}"
    application_id      = "{Azure Client Id}"
}

# AWS Hypervisor
resource "citrix_daas_hypervisor" "example-aws-hypervisor" {
    name              = "example-aws-hypervisor"
    connection_type   = "AWS"
    zone              = "{Zone Id}"
    api_key           = "{AWS account Access Key}"
    secret_key        = "{AWS account Secret Key}"
    aws_region        = "us-east-2"
}

# GCP Hypervisor
resource "citrix_daas_hypervisor" "example-gcp-hypervisor" {
    name               = "example-gcp-hypervisor"
    connection_type    = "GoogleCloudPlatform"
    zone               = "{Zone Id}"
    project_name       = "{GCP project name}"
    service_account_id = "{GCP service account Id}"
    service_account_credentials = "{GCP service account private key}"    
}