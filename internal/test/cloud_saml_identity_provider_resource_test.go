// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCSamlIdpResourcePreCheck(t *testing.T) {
	if name := os.Getenv("TEST_CC_SAML_IDP_NAME"); name == "" {
		t.Fatal("TEST_CC_SAML_IDP_NAME must be set for acceptance tests")
	}

	// Okta Idp Configurations for creation
	if okta_domain := os.Getenv("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME"); okta_domain == "" {
		t.Fatal("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME must be set for acceptance tests")
	}

	if okta_client_id := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID"); okta_client_id == "" {
		t.Fatal("TEST_CC_SAML_IDP_ENTITY_ID must be set for acceptance tests")
	}

	if okta_client_secret := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID"); okta_client_secret == "" {
		t.Fatal("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID must be set for acceptance tests")
	}

	if okta_api_token := os.Getenv("TEST_CC_SAML_IDP_SSO_URL"); okta_api_token == "" {
		t.Fatal("TEST_CC_SAML_IDP_SSO_URL must be set for acceptance tests")
	}

	// Okta Idp Configurations for update
	if okta_domain_updated := os.Getenv("TEST_CC_SAML_IDP_CERT_FILE_PATH"); okta_domain_updated == "" {
		t.Fatal("TEST_CC_SAML_IDP_CERT_FILE_PATH must be set for acceptance tests")
	}

	if okta_client_id_updated := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL"); okta_client_id_updated == "" {
		t.Fatal("TEST_CC_SAML_IDP_LOGOUT_URL must be set for acceptance tests")
	}

	if okta_client_id := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID_UPDATED"); okta_client_id == "" {
		t.Fatal("TEST_CC_SAML_IDP_ENTITY_ID_UPDATED must be set for acceptance tests")
	}

	if okta_client_secret := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID_UPDATED"); okta_client_secret == "" {
		t.Fatal("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID_UPDATED must be set for acceptance tests")
	}

	if okta_api_token := os.Getenv("TEST_CC_SAML_IDP_SSO_URL_UPDATED"); okta_api_token == "" {
		t.Fatal("TEST_CC_SAML_IDP_SSO_URL_UPDATED must be set for acceptance tests")
	}

	// Okta Idp Configurations for update
	if okta_domain_updated := os.Getenv("TEST_CC_SAML_IDP_CERT_FILE_PATH_UPDATED"); okta_domain_updated == "" {
		t.Fatal("TEST_CC_SAML_IDP_CERT_FILE_PATH_UPDATED must be set for acceptance tests")
	}

	if okta_client_id_updated := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL_UPDATED"); okta_client_id_updated == "" {
		t.Fatal("TEST_CC_SAML_IDP_LOGOUT_URL_UPDATED must be set for acceptance tests")
	}
}

func TestCCSamlIdpResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	name := os.Getenv("TEST_CC_SAML_IDP_NAME")
	authDomainName := os.Getenv("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME")
	entity_id := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID")
	use_scoped_entity_id := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID")
	single_sign_on_service_url := os.Getenv("TEST_CC_SAML_IDP_SSO_URL")
	logout_url := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL")

	name_updated := os.Getenv("TEST_CC_SAML_IDP_NAME") + "-updated"
	authDomainName_updated := os.Getenv("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME") + "Updated"
	entity_id_updated := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID_UPDATED")
	use_scoped_entity_id_updated := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID_UPDATED")
	single_sign_on_service_url_updated := os.Getenv("TEST_CC_SAML_IDP_SSO_URL_UPDATED")
	logout_url_updated := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCSamlIdpResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildCCSamlIdentityProviderResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "id"),
					// Verify the name of the Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "name", name),
					// Verify the auth domain name of the Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "auth_domain_name", authDomainName),
					// Verify SAML 2.0 Identity Provider configurations
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "entity_id", entity_id),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "use_scoped_entity_id", use_scoped_entity_id),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "single_sign_on_service_url", single_sign_on_service_url),
					resource.TestCheckResourceAttrSet("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "cert_expiration"),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "logout_url", logout_url),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_cloud_saml_identity_provider.test_saml_identity_provider",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cert_file_path"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Updated and Read testing
			{
				Config: BuildCCSamlIdentityProviderResource_Updated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "id"),
					// Verify the name of the Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "name", name_updated),
					// Verify the auth domain name of the Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "auth_domain_name", authDomainName_updated),
					// Verify SAML 2.0 Identity Provider configurations
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "entity_id", entity_id_updated),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "use_scoped_entity_id", use_scoped_entity_id_updated),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "single_sign_on_service_url", single_sign_on_service_url_updated),
					resource.TestCheckResourceAttrSet("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "cert_expiration"),
					resource.TestCheckResourceAttr("citrix_cloud_saml_identity_provider.test_saml_identity_provider", "logout_url", logout_url_updated),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	samlIdentityProviderTestResource = `
resource "citrix_cloud_saml_identity_provider" "test_saml_identity_provider" {
	name                                = "%s"
    auth_domain_name                    = "%s"
    
    entity_id                           = "%s"
    use_scoped_entity_id                = %s

    single_sign_on_service_url          = "%s"
    sign_auth_request                   = true
    single_sign_on_service_binding      = "HttpPost"
    saml_response                       = "SignEitherResponseOrAssertion"
    cert_file_path                      = "%s"
    authentication_context              = "Unspecified"
    authentication_context_comparison   = "Exact"

    logout_url                          = "%s"
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
`
)

func BuildCCSamlIdentityProviderResource(t *testing.T) string {
	name := os.Getenv("TEST_CC_SAML_IDP_NAME")
	auth_domain_name := os.Getenv("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME")
	entity_id := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID")
	use_scoped_entity_id := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID")
	single_sign_on_service_url := os.Getenv("TEST_CC_SAML_IDP_SSO_URL")
	cert_file_path := os.Getenv("TEST_CC_SAML_IDP_CERT_FILE_PATH")
	logout_url := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL")
	return fmt.Sprintf(samlIdentityProviderTestResource, name, auth_domain_name, entity_id, use_scoped_entity_id, single_sign_on_service_url, cert_file_path, logout_url)
}

func BuildCCSamlIdentityProviderResource_Updated(t *testing.T) string {
	name := os.Getenv("TEST_CC_SAML_IDP_NAME") + "-updated"
	auth_domain_name := os.Getenv("TEST_CC_SAML_IDP_AUTH_DOMAIN_NAME") + "Updated"
	entity_id := os.Getenv("TEST_CC_SAML_IDP_ENTITY_ID_UPDATED")
	use_scoped_entity_id := os.Getenv("TEST_CC_SAML_IDP_USE_SCOPED_ENTITY_ID_UPDATED")
	single_sign_on_service_url := os.Getenv("TEST_CC_SAML_IDP_SSO_URL_UPDATED")
	cert_file_path := os.Getenv("TEST_CC_SAML_IDP_CERT_FILE_PATH_UPDATED")
	logout_url := os.Getenv("TEST_CC_SAML_IDP_LOGOUT_URL_UPDATED")
	return fmt.Sprintf(samlIdentityProviderTestResource, name, auth_domain_name, entity_id, use_scoped_entity_id, single_sign_on_service_url, cert_file_path, logout_url)
}
