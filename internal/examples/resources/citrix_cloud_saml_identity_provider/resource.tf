resource "citrix_cloud_saml_identity_provider" "example_saml_idp" {
    name                                = "example SAML 2.0 idp"
    auth_domain_name                    = "exampleAuthDomain"
    
    entity_id                           = var.saml_entity_id
    use_scoped_entity_id                = false

    single_sign_on_service_url          = var.saml_sso_url
    sign_auth_request                   = true
    single_sign_on_service_binding      = "HttpPost"
    saml_response                       = "SignEitherResponseOrAssertion"
    cert_file_path                      = var.cert_file_path
    authentication_context              = "Unspecified"
    authentication_context_comparison   = "Exact"

    logout_url                          = var.saml_logout_url
    sign_logout_request                 = true
    logout_binding                      = "HttpPost"

    attribute_names = {
        user_display_name       = "displayName"
        user_given_name         = "givenName"
        user_family_name        = "familyName"
        security_identifier     = "cip_sid"
        user_principal_name     = "cip_upn"
        email                   = "cip_email"
        ad_object_identifier    = "cip_oid"
        ad_forest               = "cip_forest"
        ad_domain               = "cip_domain"
    }
}
