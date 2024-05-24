// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment

func TestSTFStoreServicePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH"); v == "" {
		t.Fatal("TEST_STF_STORE_VIRTUAL_PATH must be set for acceptance tests")
	}
}

func TestSTFStoreServiceResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFStoreServicePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFStoreServiceResource(t, testSTFStoreServiceResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "site_id", siteId),
					// Verify virtual_path of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "virtual_path", virtualPath),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_store_service.testSTFStoreService",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "virtual_path",
				ImportStateIdFunc:                    generateImportStateId_STFStoreService,
				ImportStateVerifyIgnore:              []string{"last_updated", "authentication_service"},
			},

			// Update testing for STF Store Service
			{
				Config: BuildSTFStoreServiceResource(t, testSTFStoreServiceResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify friendly_name of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "friendly_name", "Store_Updated"),
				),
			},
		},
	})
}

func BuildSTFStoreServiceResource(t *testing.T, storeService string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH")

	return BuildSTFAuthenticationServiceResource(t, testSTFAuthenticationServiceResources) + fmt.Sprintf(storeService, siteId, virtualPath)

}

func generateImportStateId_STFStoreService(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_store_service.testSTFStoreService"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["site_id"], rawState["virtual_path"]), nil
}

var (
	testSTFStoreServiceResources = `
	resource "citrix_stf_store_service" "testSTFStoreService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "Store"
		authentication_service =  citrix_stf_authentication_service.testSTFAuthenticationService.virtual_path
	  }
	`
	testSTFStoreServiceResources_updated = `
	resource "citrix_stf_store_service" "testSTFStoreService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "Store_Updated"
		authentication_service =  citrix_stf_authentication_service.testSTFAuthenticationService.virtual_path
	  }
	`
)
