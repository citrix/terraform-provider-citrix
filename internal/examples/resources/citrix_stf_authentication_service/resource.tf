resource "citrix_stf_authentication_service" "example-stf-authentication-service" {
  site_id             = citrix_stf_deployment.example-stf-deployment.site_id
  friendly_name       = "Example STF Authentication Service"
  virtual_path        = "/Citrix/Authentication"
  claims_factory_name = "ExampleClaimsFactoryName"

  // Add depends_on attribute to ensure the StoreFront Authentication Service is created after the Deployment
  depends_on = [ citrix_stf_deployment.example-stf-deployment ]
}