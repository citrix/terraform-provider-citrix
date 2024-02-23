# Nutanix Hypervisor
resource "citrix_nutanix_hypervisor" "example-nutanix-hypervisor" {
    name               = "example-nutanix-hypervisor"
    zone               = "<Zone Id>"
    username           = "<Username>"
    password           = "<Password_In_Plaintext>"
    password_format    = "Plaintext"
    addresses          = ["10.122.36.26"]
    max_absolute_active_actions = 20
}