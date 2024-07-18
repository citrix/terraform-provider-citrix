package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSTFXenappDefaultStorePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH"); v == "" {
		t.Fatal("TEST_STF_STORE_VIRTUAL_PATH must be set for acceptance tests")
	}
}

func TestSTFXenappDefaultStoreResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFStoreServicePreCheck(t)
			TestSTFAuthenticationServicePreCheck(t)
			TestSTFXenappDefaultStorePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFXenappDefaultStoreResource(t, testSTFXenappDefaultStoreResources),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the STF Default Store
					resource.TestCheckResourceAttr("citrix_stf_xenapp_default_store.testSTFXenappDefaultStore", "store_site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_xenapp_default_store.testSTFXenappDefaultStore", "store_virtual_path", virtualPath),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_xenapp_default_store.testSTFXenappDefaultStore",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "farm_name",
				ImportStateIdFunc:                    generateImportStateId_STFXenappDefaultStore,
				ImportStateVerifyIgnore:              []string{"last_updated"},
			},

			// Update testing for STF authentication service
			{
				Config: BuildSTFXenappDefaultStoreResource(t, testSTFXenappDefaultStoreResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the STF Default Store
					resource.TestCheckResourceAttr("citrix_stf_xenapp_default_store.testSTFXenappDefaultStore", "store_site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_xenapp_default_store.testSTFXenappDefaultStore", "store_virtual_path", virtualPath+"updated"),
				),
			},
		},
	})
}

func BuildSTFXenappDefaultStoreResource(t *testing.T, XenappDefaultStore string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_STORE_VIRTUAL_PATH")
	return BuildSTFStoreServiceResource(t, testSTFStoreServiceResources) + fmt.Sprintf(XenappDefaultStore, siteId, virtualPath)
}

func generateImportStateId_STFXenappDefaultStore(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_xenapp_default_store.testSTFXenappDefaultStore"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["store_site_id"], rawState["store_virtual_path"]), nil
}

var (
	testSTFXenappDefaultStoreResources = `
	resource "citrix_stf_xenapp_default_store" "testSTFXenappDefaultStore" {
		store_site_id = "%s"
		store_virtual_path      = "%s"
	}
	`

	testSTFXenappDefaultStoreResources_updated = `
	resource "citrix_stf_store_service" "testSTFStoreServiceUpdated" {
		site_id       = "%s"
		virtual_path = "%supdated"
		friendly_name = "Store"
		authentication_service_virtual_path =  citrix_stf_authentication_service.testSTFAuthenticationService.virtual_path
	  }


    resource "citrix_stf_xenapp_default_store" "testSTFXenappDefaultStore" {
		store_virtual_path      = citrix_stf_store_service.testSTFStoreServiceUpdated.virtual_path
		store_site_id = citrix_stf_store_service.testSTFStoreService.site_id
	}
	`
)
