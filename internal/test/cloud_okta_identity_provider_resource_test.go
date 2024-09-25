// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCOktaIdpResourcePreCheck(t *testing.T) {
	if name := os.Getenv("TEST_CC_OKTA_IDP_NAME"); name == "" {
		t.Fatal("TEST_CC_OKTA_IDP_NAME must be set for acceptance tests")
	}

	// Okta Idp Configurations for creation
	if okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN"); okta_domain == "" {
		t.Fatal("TEST_CC_OKTA_IDP_DOMAIN must be set for acceptance tests")
	}

	if okta_client_id := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_ID"); okta_client_id == "" {
		t.Fatal("TEST_CC_OKTA_IDP_CLIENT_ID must be set for acceptance tests")
	}

	if okta_client_secret := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_SECRET"); okta_client_secret == "" {
		t.Fatal("TEST_CC_OKTA_IDP_CLIENT_SECRET must be set for acceptance tests")
	}

	if okta_api_token := os.Getenv("TEST_CC_OKTA_IDP_API_TOKEN"); okta_api_token == "" {
		t.Fatal("TEST_CC_OKTA_IDP_API_TOKEN must be set for acceptance tests")
	}

	// Okta Idp Configurations for update
	if okta_domain_updated := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN_UPDATED"); okta_domain_updated == "" {
		t.Fatal("TEST_CC_OKTA_IDP_DOMAIN_UPDATED must be set for acceptance tests")
	}

	if okta_client_id_updated := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_ID_UPDATED"); okta_client_id_updated == "" {
		t.Fatal("TEST_CC_OKTA_IDP_CLIENT_ID_UPDATED must be set for acceptance tests")
	}

	if okta_client_secret_updated := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_SECRET_UPDATED"); okta_client_secret_updated == "" {
		t.Fatal("TEST_CC_OKTA_IDP_CLIENT_SECRET_UPDATED must be set for acceptance tests")
	}

	if okta_api_token_updated := os.Getenv("TEST_CC_OKTA_IDP_API_TOKEN_UPDATED"); okta_api_token_updated == "" {
		t.Fatal("TEST_CC_OKTA_IDP_API_TOKEN_UPDATED must be set for acceptance tests")
	}
}

func TestCCOktaIdpResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	name := os.Getenv("TEST_CC_OKTA_IDP_NAME")
	okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN")

	name_updated := os.Getenv("TEST_CC_OKTA_IDP_NAME") + "-updated"
	okta_domain_updated := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCOktaIdpResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildCCOktaIdentityProviderResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Okta Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id"),
					// Verify the name of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name),
					// Verify the domain of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_cloud_okta_identity_provider.test_okta_identity_provider",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"okta_client_id", "okta_client_secret", "okta_api_token"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Testing Update Name Only
			{
				Config: BuildCCOktaIdentityProviderResource_NameUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Okta Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id"),
					// Verify the name of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name_updated),
					// Verify the domain of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Testing Update Okta Configs Only
			{
				Config: BuildCCOktaIdentityProviderResource_NameAndConfigsUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Okta Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id"),
					// Verify the name of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name_updated),
					// Verify the domain of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain_updated),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Testing Update Okta Name and Configs
			{
				Config: BuildCCOktaIdentityProviderResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Okta Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id"),
					// Verify the name of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name),
					// Verify the domain of the Okta Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	oktaIdentityProviderTestResource = `
resource "citrix_cloud_okta_identity_provider" "test_okta_identity_provider" {
	name 				= "%s"
	okta_domain 		= "%s"
	okta_client_id 		= "%s"
	okta_client_secret	= "%s"
	okta_api_token 		= "%s"
}
`
)

func BuildCCOktaIdentityProviderResource(t *testing.T) string {
	name := os.Getenv("TEST_CC_OKTA_IDP_NAME")
	okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN")
	okta_client_id := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_ID")
	okta_client_secret := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_SECRET")
	okta_api_token := os.Getenv("TEST_CC_OKTA_IDP_API_TOKEN")
	return fmt.Sprintf(oktaIdentityProviderTestResource, name, okta_domain, okta_client_id, okta_client_secret, okta_api_token)
}

func BuildCCOktaIdentityProviderResource_NameUpdated(t *testing.T) string {
	name := os.Getenv("TEST_CC_OKTA_IDP_NAME") + "-updated"
	okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN")
	okta_client_id := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_ID")
	okta_client_secret := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_SECRET")
	okta_api_token := os.Getenv("TEST_CC_OKTA_IDP_API_TOKEN")
	return fmt.Sprintf(oktaIdentityProviderTestResource, name, okta_domain, okta_client_id, okta_client_secret, okta_api_token)
}

func BuildCCOktaIdentityProviderResource_NameAndConfigsUpdated(t *testing.T) string {
	name := os.Getenv("TEST_CC_OKTA_IDP_NAME") + "-updated"
	okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DOMAIN_UPDATED")
	okta_client_id := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_ID_UPDATED")
	okta_client_secret := os.Getenv("TEST_CC_OKTA_IDP_CLIENT_SECRET_UPDATED")
	okta_api_token := os.Getenv("TEST_CC_OKTA_IDP_API_TOKEN_UPDATED")
	return fmt.Sprintf(oktaIdentityProviderTestResource, name, okta_domain, okta_client_id, okta_client_secret, okta_api_token)
}
