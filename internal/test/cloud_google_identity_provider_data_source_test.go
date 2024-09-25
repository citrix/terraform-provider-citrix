// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCGoogleIdpDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_DATA_SOURCE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_AUTH_DOMAIN_NAME"); v == "" {
		t.Fatal("TEST_CC_GOOGLE_IDP_DATA_SOURCE_AUTH_DOMAIN_NAME must be set for acceptance tests")
	}
}

func TestCCGoogleIdpDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_ID")
	name := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_NAME")
	authDomainName := os.Getenv("TEST_CC_GOOGLE_IDP_DATA_SOURCE_AUTH_DOMAIN_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCGoogleIdpDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildCCGoogleIdentityProviderDataSource(t, cc_google_idp_test_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildCCGoogleIdentityProviderDataSource(t, cc_google_idp_test_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_google_identity_provider.test_google_identity_provider", "auth_domain_name", authDomainName),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildCCGoogleIdentityProviderDataSource(t *testing.T, googleIdpDataSource string, idOrName string) string {
	return fmt.Sprintf(googleIdpDataSource, idOrName)
}

var (
	cc_google_idp_test_data_source_using_id = `
	data "citrix_cloud_google_identity_provider" "test_google_identity_provider" {
		id         = "%s"
	}
	`

	cc_google_idp_test_data_source_using_name = `
	data "citrix_cloud_google_identity_provider" "test_google_identity_provider" {
		name       = "%s"
	}
	`
)
