resource "citrix_vsphere_hypervisor" "example-vsphere-hypervisor" {
    name            = "example-vsphere-hyperv"
    zone            = citrix_zone.example-zone.id
    username        = "<vSphere Username>"
    password        = "<vSphere Password>"        
    password_format = "PlainText"
    addresses       = [
        "http://<IP address or hostname for vSphere>"
    ]
    max_absolute_active_actions = 20
}