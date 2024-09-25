# Get Citrix Cloud Google Identity Provider data source by ID
data "citrix_cloud_google_identity_provider" "example_google_identity_provider" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get Citrix Cloud Google Identity Provider data source by name
data "citrix_cloud_google_identity_provider" "example_google_identity_provider" {
  name = "exampleGoogleIdentityProvider"
}
