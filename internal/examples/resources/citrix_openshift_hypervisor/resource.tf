resource "citrix_openshift_hypervisor" "example-openshift-hypervisor" {
    name                                     = "example-openshift-hypervisor"
    zone                                     = "<Zone Id>"
    service_account_token                    = "<Service_Access_Token_In_Plaintext>"
    addresses                                = ["<Openshift_cluster_connection_address>"]
    ssl_thumbprints                          = ["<SSL_Thumbprint>"]
    max_absolute_active_actions              = 150
    max_absolute_new_actions_per_minute      = 30
    max_power_actions_percentage_of_machines = 40
}