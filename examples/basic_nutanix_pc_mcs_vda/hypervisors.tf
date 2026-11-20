resource "citrix_nutanix_pc_hypervisor" "example-nutanix-pc-hypervisor" {
    name            = var.hypervisor_name
    zone            = citrix_zone.example-zone.id
    username        = var.nutanix_pc_username
    password        = var.nutanix_pc_password
    password_format = var.nutanix_pc_password_format
    addresses       = var.nutanix_pc_addresses
    ssl_thumbprints = var.nutanix_pc_ssl_thumbprints
}
