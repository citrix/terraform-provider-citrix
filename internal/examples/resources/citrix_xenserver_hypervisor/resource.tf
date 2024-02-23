# XenServer Hypervisor
resource "citrix_xenserver_hypervisor" "example-xenserver-hypervisor" {
    name            = "example-xenserver-hypervisor"
    zone            = "<Zone Id>"
    username        = "<XenServer username>"
    password        = "<XenServer password>"        
    password_format = "PlainText"
    addresses       = [
        "http://<IP address or hostname for XenServer>"
    ]
    ssl_thumbprints = [
        "<SSL thumbprint in SHA-256 format without colons>"
    ]
}