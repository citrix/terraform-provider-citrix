resource "citrix_cloud_admin_user" "example-full-admin-user" {
  access_type   = "Full"
  email         = "example-full-admin@citrix.com"
  provider_type = "CitrixSts"
  type          = "AdministratorUser"
}

resource "citrix_cloud_admin_user" "example-custom-admin-user" {
  access_type   = "Custom"
  email         = "example-custom-admin@citrix.com"
  provider_type = "CitrixSts"
  type          = "AdministratorUser"
  policies = [
    {
      name         = "Example Policy 1"
      service_name = "XenDesktop"
      scopes       = ["Scope1", "Scope2"]
    },
    {
      name = "Example Policy 2"
    }
  ]
}
