# Vsphere Hypervisor
resource "citrix_vsphere_hypervisor" "example-vsphere-hypervisor" {
    name               = "example-vsphere-hypervisor"
    zone               = "<Zone Id>"
    username           = "<Username>"
    password           = "<Password_In_Plaintext>"
    password_format    = "Plaintext"
    addresses          = ["https://10.36.122.45"]
    max_absolute_active_actions = 20
}