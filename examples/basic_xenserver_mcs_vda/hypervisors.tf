resource "citrix_xenserver_hypervisor" "example-xenserver-hypervisor" {
    name            = "example-xenserver-hyperv"
    zone            = citrix_zone.example-zone.id
    username        = "<XenServer Username>"
    password        = "<XenServer Password>"        
    password_format = "PlainText"
    addresses       = [
        "http://<IP address or hostname for XenServer>"
    ]
    ssl_thumbprints = [
        "<SSL thumbprint in SHA-256 format without colons>"
    ]
}