resource "citrix_cloud_google_identity_provider" "example_google_idp" {
    name              = "example Google idp"
    auth_domain_name  = "exampleAuthDomain"
    client_email      = var.example_google_idp_client_email
    private_key       = var.example_google_idp_private_key
    impersonated_user = var.example_google_idp_impersonated_user
}
