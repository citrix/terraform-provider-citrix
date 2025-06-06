# Remote PC Wake on LAN Hypervisor
resource "citrix_remote_pc_wake_on_lan_hypervisor" "example-remotepc-wakeonlan-hypervisor" {
    name                = "example-remotepc-wakeonlan-hypervisor"
    zone                = "<Zone Id>"
}
