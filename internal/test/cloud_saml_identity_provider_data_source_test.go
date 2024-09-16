// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCCSamlIdpDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_CC_SAML_IDP_DATA_SOURCE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_CC_SAML_IDP_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_AUTH_DOMAIN"); v == "" {
		t.Fatal("TEST_CC_SAML_IDP_DATA_SOURCE_AUTH_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_ENTITY_ID"); v == "" {
		t.Fatal("TEST_CC_SAML_IDP_DATA_SOURCE_ENTITY_ID must be set for acceptance tests")
	}
}

func TestCCSamlIdpDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_ID")
	name := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_NAME")
	authDomain := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_AUTH_DOMAIN")
	entityId := os.Getenv("TEST_CC_SAML_IDP_DATA_SOURCE_ENTITY_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCCSamlIdpDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildCCSamlIdentityProviderDataSource(t, cc_saml_idp_test_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "auth_domain_name", authDomain),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "entity_id", entityId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildCCSamlIdentityProviderDataSource(t, cc_saml_idp_test_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "id", id),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "name", name),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "auth_domain_name", authDomain),
					resource.TestCheckResourceAttr("data.citrix_cloud_saml_identity_provider.test_saml_identity_provider", "entity_id", entityId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildCCSamlIdentityProviderDataSource(t *testing.T, samlIdpDataSource string, idOrName string) string {
	return fmt.Sprintf(samlIdpDataSource, idOrName)
}

var (
	cc_saml_idp_test_data_source_using_id = `
	data "citrix_cloud_saml_identity_provider" "test_saml_identity_provider" {
		id         = "%s"
	}
	`

	cc_saml_idp_test_data_source_using_name = `
	data "citrix_cloud_saml_identity_provider" "test_saml_identity_provider" {
		name       = "%s"
	}
	`
)
