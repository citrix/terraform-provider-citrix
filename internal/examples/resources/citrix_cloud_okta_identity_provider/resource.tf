resource "citrix_cloud_okta_identity_provider" "example_okta_idp" {
    name               = "example Okta idp"
    okta_domain        = "example.okta.com"
    okta_client_id     = var.example_okta_client_id
    okta_client_secret = var.example_okta_client_secret
    okta_api_token     = var.example_okta_api_token
}
