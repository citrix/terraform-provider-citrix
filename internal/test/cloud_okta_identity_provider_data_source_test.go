// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCOktaIdpDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_CC_OKTA_IDP_DATA_SOURCE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_CC_OKTA_IDP_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_DOMAIN"); v == "" {
		t.Fatal("TEST_CC_OKTA_IDP_DATA_SOURCE_DOMAIN must be set for acceptance tests")
	}
}

func TestCCOktaIdpDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_ID")
	name := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_NAME")
	okta_domain := os.Getenv("TEST_CC_OKTA_IDP_DATA_SOURCE_DOMAIN")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCOktaIdpDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildCCOktaIdentityProviderDataSource(t, cc_okta_idp_test_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildCCOktaIdentityProviderDataSource(t, cc_okta_idp_test_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_okta_identity_provider.test_okta_identity_provider", "okta_domain", okta_domain),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildCCOktaIdentityProviderDataSource(t *testing.T, oktaIdpDataSource string, idOrName string) string {
	return fmt.Sprintf(oktaIdpDataSource, idOrName)
}

var (
	cc_okta_idp_test_data_source_using_id = `
	data "citrix_cloud_okta_identity_provider" "test_okta_identity_provider" {
		id         = "%s"
	}
	`

	cc_okta_idp_test_data_source_using_name = `
	data "citrix_cloud_okta_identity_provider" "test_okta_identity_provider" {
		name       = "%s"
	}
	`
)
