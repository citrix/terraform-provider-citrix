resource "citrix_stf_roaming_gateway" "example-stf-roaming-gateway" {
    site_id                        = citrix_stf_deployment.example-stf-deployment.site_id
    name                           = "Example Roaming Gateway Name"
    logon_type                     = "Domain"
    smart_card_fallback_logon_type = "None"
    gateway_url                    = "https://example.gateway.url/"
    callback_url                   = "https://example.callback.url/"
    version                        = "Version10_0_69_4"
    subnet_ip_address              = "10.0.0.1"
    stas_bypass_duration           = "0.1:0:0"
    gslb_url                       = "https://example.gslb.url/"
    session_reliability            = false
    request_ticket_two_stas        = false
    stas_use_load_balancing        = false
    is_cloud_gateway               = false
    secure_ticket_authority_urls = [ {
      sta_url = "https://example.url/scripts/ctxsta.dll"
      sta_validation_enabled = true
      sta_validation_secret = "<STA Validation Secret>"
    }]

    // Add depends_on attribute to ensure the StoreFront Roaming Gateway is created after the Deployment
    depends_on = [ citrix_stf_deployment.example-stf-deployment ]
}
