resource "citrix_xenserver_hypervisor" "example-xenserver-hypervisor" {
    name            = var.hypervisor_name
    zone            = citrix_zone.example-zone.id
    username        = var.xenserver_username
    password        = var.xenserver_password    
    password_format = var.xenserver_password_format
    addresses       = var.xenserver_addresses
}