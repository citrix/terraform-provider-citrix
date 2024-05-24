resource "citrix_stf_authentication_service" "example-stf-authentication-service" {
  site_id       = "1"
  friendly_name = "Example STF Authentication Service"
  virtual_path  = "/Citrix/Authentication"
}