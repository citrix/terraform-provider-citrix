# HPE Moonshot Hypervisor
resource "citrix_hpe_moonshot_hypervisor" "example-hpe-moonshot-hypervisor" {
    name            = "example-hpe-moonshot-hypervisor"
    zone            = "<Zone Id>"
    username        = "<HPE Moonshot username>"
    password        = "<HPE Moonshot password>"        
    password_format = "PlainText"
    addresses       = [
        "http://<IP address or hostname for HPE Moonshot>"
    ]
    ssl_thumbprints = [
        "<SSL thumbprint in SHA-256 format without colons>"
    ]
}