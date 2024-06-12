// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestSTFWebReceiverServicePreCheck validates the necessary test API keys exist in the testing environment

func TestSTFWebReceiverServicePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH"); v == "" {
		t.Fatal("TEST_STF_WEBRECEIVER_VIRTUAL_PATH must be set for acceptance tests")
	}
}

func TestSTFWebReceiverServiceResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFWebReceiverServicePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFWebReceiverServiceResource(t, testSTFWebReceiverServiceResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "site_id", siteId),
					// Verify virtual_path of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "virtual_path", virtualPath),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_webreceiver_service.testSTFWebReceiverService",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "virtual_path",
				ImportStateIdFunc:                    generateImportStateId_STFWebReceiverService,
				ImportStateVerifyIgnore:              []string{"last_updated", "store_service"},
			},

			// Update testing for STF WebReceiver Service
			{
				Config: BuildSTFWebReceiverServiceResource(t, testSTFWebReceiverServiceResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify friendly_name of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "friendly_name", "WebReceiver_Updated"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "authentication_methods.#", "2"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "plugin_assistant.enabled", "true"),
				),
			},
		},
	})
}

func BuildSTFWebReceiverServiceResource(t *testing.T, webreceiverService string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH")

	return BuildSTFStoreServiceResource(t, testSTFStoreServiceResources) + fmt.Sprintf(webreceiverService, siteId, virtualPath)

}

func generateImportStateId_STFWebReceiverService(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_webreceiver_service.testSTFWebReceiverService"
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
	testSTFWebReceiverServiceResources = `
	resource "citrix_stf_webreceiver_service" "testSTFWebReceiverService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "WebReceiver"
		store_service = citrix_stf_store_service.testSTFStoreService.virtual_path
	
	  }
	`
	testSTFWebReceiverServiceResources_updated = `
	resource "citrix_stf_webreceiver_service" "testSTFWebReceiverService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "WebReceiver_Updated" 
		store_service = citrix_stf_store_service.testSTFStoreService.virtual_path
		authentication_methods = [ 
			"ExplicitForms", 
			"CitrixAGBasic"
		  ]
		plugin_assistant = {
			enabled = true
			html5_single_tab_launch = true
		}
	  }
	`
)
