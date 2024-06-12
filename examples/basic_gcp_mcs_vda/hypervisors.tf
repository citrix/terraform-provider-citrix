resource "citrix_gcp_hypervisor" "example-gcp-hypervisor" {
    name                = var.hypervisor_name
    zone                = citrix_zone.example-zone.id
    service_account_id  = var.gcp_service_account_id
    service_account_credentials = var.gcp_service_account_credentials
}

