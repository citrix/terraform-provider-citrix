resource "citrix_gcp_hypervisor" "example-gcp-hypervisor" {
    name                = "example-gcp-hyperv"
    zone                = citrix_zone.example-zone.id
    service_account_id = "{GCP service account Id}"
    service_account_credentials = "{GCP service account private key}"    
}

