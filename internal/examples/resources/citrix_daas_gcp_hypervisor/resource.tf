# GCP Hypervisor
resource "citrix_daas_gcp_hypervisor" "example-gcp-hypervisor" {
    name               = "example-gcp-hypervisor"
    zone               = "{Zone Id}"
    service_account_id = "{GCP service account Id}"
    service_account_credentials = "{GCP service account private key}"    
}