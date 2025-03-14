resource "citrix_site_settings" "example" {
    web_ui_policy_set_enabled                           = false
    dns_resolution_enabled                              = false
    multiple_remote_pc_assignments                      = true
    trust_requests_sent_to_the_xml_service_port_enabled = false
}
