resource "citrix_nutanix_hypervisor" "example-nutanix-hypervisor" {
    name            = "example-nutanix-hyperv"
    zone            = citrix_zone.example-zone.id
    username        = "<Username>"
    password        = "<Password>"        
    password_format = "PlainText"
    addresses       = [
        "http://<IP address or hostname for Nutanix>"
    ]
}