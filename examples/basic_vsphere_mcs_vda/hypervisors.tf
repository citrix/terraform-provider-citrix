resource "citrix_vsphere_hypervisor" "example-vsphere-hypervisor" {
    name            = var.hypervisor_name
    zone            = citrix_zone.example-zone.id
    username        = var.vsphere_username
    password        = var.vsphere_password
    password_format = var.vsphere_password_format
    addresses       = var.vsphere_addresses
    max_absolute_active_actions = 20
}