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

resource "citrix_cloud_admin_user" "example-custom-azure-ad-admin-group" {
  access_type          = "Custom"
  provider_type        = "AzureAd"
  display_name         = "Example Custom Azure Ad Admin Group"
  type                 = "AdministratorGroup"
  external_provider_id = "Example Azure Tenant Id"
  external_user_id     = "Example Azure Group Id"
  policies = [
    {
      name         = "Example Policy 1"
      scopes       = ["Scope1", "Scope2"]
    },
    {
      name = "Example Policy 2"
    }
  ]
}

resource "citrix_cloud_admin_user" "example-custom-ad-admin-group" {
  access_type          = "Custom"
  provider_type        = "Ad"
  display_name         = "Example Custom AD Admin Group"
  type                 = "AdministratorGroup"
  external_provider_id = "<DomainFQDN>"
  external_user_id     = "Example Group Id"
  policies = [
    {
      name         = "Example Policy 1"
    }
  ]
}