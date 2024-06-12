resource "citrix_azure_hypervisor" "example-azure-hypervisor" {
    name                = var.hypervisor_name
    zone                = citrix_zone.example-zone.id
    application_id      = var.azure_application_id
    application_secret  = var.azure_application_secret
    subscription_id     = var.azure_subscription_id
    active_directory_id = var.azure_tenant_id
}