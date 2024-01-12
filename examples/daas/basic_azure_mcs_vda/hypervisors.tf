resource "citrix_daas_azure_hypervisor" "example-azure-hypervisor" {
    name                = "example-azure-hyperv"
    zone                = citrix_daas_zone.example-zone.id
    application_id      = "<Azure SPN client ID>"
    application_secret  = "<Azure SPN client secret>"
    subscription_id     = "<Azure Subscription ID>"
    active_directory_id = "<Azure Tenant ID>"
}