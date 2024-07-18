resource "citrix_cloud_admin_user" "example-admin-user" {
  access_type = "Full"
  email = "example-admin@citrix.com"
  provider_type = "CitrixSts"
  type = "AdministratorUser"
}