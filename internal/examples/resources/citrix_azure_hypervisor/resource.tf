# Azure Hypervisor
resource "citrix_azure_hypervisor" "example-azure-hypervisor" {
    name                = "example-azure-hypervisor"
    zone                = "<Zone Id>"
    active_directory_id = "<Azure Tenant Id>"
    subscription_id     = "<Azure Subscription Id>"
    application_secret  = var.azure_client_secret # Azure Client Secret from variable
    application_id      = "<Azure Client Id>"
}