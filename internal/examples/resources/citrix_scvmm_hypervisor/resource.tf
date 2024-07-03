# SCVMM Hypervisor
resource "citrix_scvmm_hypervisor" "example-scvmm-hypervisor" {
    name               = "example-scvmm-hypervisor"
    zone               = "<Zone Id>"
    username           = "<Username>"
    password           = "<Password_In_Plaintext>"
    password_format    = "Plaintext"
    addresses          = ["scvmm.example.com"]
    max_absolute_active_actions = 50
}