# Nutanix Prism Central Hypervisor
resource "citrix_nutanix_pc_hypervisor" "example-nutanix-pc-hypervisor" {
    name               = "example-nutanix-pc-hypervisor"
    zone               = "<Zone Id>"
    username           = "<Username>"
    password           = "<Password_In_Plaintext>"
    password_format    = "PlainText"
    addresses          = ["10.122.36.30"]
    ssl_thumbprints    = ["<SSL_Thumbprint>"]
}
