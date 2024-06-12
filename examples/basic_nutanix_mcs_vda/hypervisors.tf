resource "citrix_nutanix_hypervisor" "example-nutanix-hypervisor" {
    name            = var.hypervisor_name
    zone            = citrix_zone.example-zone.id
    username        = var.nutanix_username
    password        = var.nutanix_password
    password_format = var.nutanix_password_format
    addresses       = var.nutanix_addresses
}