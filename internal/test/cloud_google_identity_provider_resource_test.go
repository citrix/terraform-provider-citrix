// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCGoogleIdpResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_NAME"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME must be set for acceptance tests")
	}

	// Google Idp Configurations for creation
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_CC_GOOGLE_IDP_PRIVATE_KEY"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_PRIVATE_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER must be set for acceptance tests")
	}

	// Google Idp Configurations for update
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL_UPDATED"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_CC_GOOGLE_IDP_PRIVATE_KEY_UPDATED"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_PRIVATE_KEY_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER_UPDATED"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER_UPDATED must be set for acceptance tests")
	}
}

func TestCCGoogleIdpResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	name := os.Getenv("TEST_CC_GOOGLE_IDP_NAME")
	authDomainName := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME")

	name_updated := os.Getenv("TEST_CC_GOOGLE_IDP_NAME") + "-updated"
	authDomainName_updated := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME") + "Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCGoogleIdpResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildCCGoogleIdentityProviderResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Google Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_google_identity_provider.test_google_identity_provider", "id"),
					// Verify the name of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name),
					// Verify the auth domain name of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_cloud_google_identity_provider.test_google_identity_provider",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_email", "private_key", "impersonated_user"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Testing Update Name and Auth Domain Name Only
			{
				Config: BuildCCGoogleIdentityProviderResource_NameUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Google Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_google_identity_provider.test_google_identity_provider", "id"),
					// Verify the name of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name_updated),
					// Verify the domain of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName_updated),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Testing Update Google Configs Only
			{
				Config: BuildCCGoogleIdentityProviderResource_NameAndConfigsUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Google Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_google_identity_provider.test_google_identity_provider", "id"),
					// Verify the name of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name_updated),
					// Verify the domain of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName_updated),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Testing Update Google Name and Configs
			{
				Config: BuildCCGoogleIdentityProviderResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the Google Identity Provider resource
					resource.TestCheckResourceAttrSet("citrix_cloud_google_identity_provider.test_google_identity_provider", "id"),
					// Verify the name of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name),
					// Verify the domain of the Google Identity Provider resource
					resource.TestCheckResourceAttr("citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	googleIdentityProviderTestResource = `
resource "citrix_cloud_google_identity_provider" "test_google_identity_provider" {
	name 				= "%s"
	auth_domain_name 	= "%s"
	client_email 		= "%s"
	private_key			= "%s"
	impersonated_user 	= "%s"
}
`
)

func BuildCCGoogleIdentityProviderResource(t *testing.T) string {
	name := os.Getenv("TEST_CC_GOOGLE_IDP_NAME")
	authDomainName := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME")
	clientEmail := os.Getenv("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL")
	privateKey := os.Getenv("TEST_CC_GOOGLE_IDP_PRIVATE_KEY")
	impersonatedUser := os.Getenv("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER")
	return fmt.Sprintf(googleIdentityProviderTestResource, name, authDomainName, clientEmail, privateKey, impersonatedUser)
}

func BuildCCGoogleIdentityProviderResource_NameUpdated(t *testing.T) string {
	name := os.Getenv("TEST_CC_GOOGLE_IDP_NAME") + "-updated"
	authDomainName := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME") + "Updated"
	clientEmail := os.Getenv("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL")
	privateKey := os.Getenv("TEST_CC_GOOGLE_IDP_PRIVATE_KEY")
	impersonatedUser := os.Getenv("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER")
	return fmt.Sprintf(googleIdentityProviderTestResource, name, authDomainName, clientEmail, privateKey, impersonatedUser)
}

func BuildCCGoogleIdentityProviderResource_NameAndConfigsUpdated(t *testing.T) string {
	name := os.Getenv("TEST_CC_GOOGLE_IDP_NAME") + "-updated"
	authDomainName := os.Getenv("TEST_CC_GOOGLE_IDP_AUTH_DOMAIN_NAME") + "Updated"
	clientEmail := os.Getenv("TEST_CC_GOOGLE_IDP_CLIENT_EMAIL_UPDATED")
	privateKey := os.Getenv("TEST_CC_GOOGLE_IDP_PRIVATE_KEY")
	impersonatedUser := os.Getenv("TEST_CC_GOOGLE_IDP_IMPERSONATED_USER_UPDATED")
	return fmt.Sprintf(googleIdentityProviderTestResource, name, authDomainName, clientEmail, privateKey, impersonatedUser)
}
