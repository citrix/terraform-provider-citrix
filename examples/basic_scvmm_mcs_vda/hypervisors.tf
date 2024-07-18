resource "citrix_scvmm_hypervisor" "example-scvmm-hypervisor" {
    name               = var.hypervisor_name
    zone               = citrix_zone.example-zone.id
    username           = var.scvmm_username
    password           = var.scvmm_password
    password_format    = var.scvmm_password_format
    addresses          = var.scvmm_addresses
    max_absolute_active_actions = 50
}