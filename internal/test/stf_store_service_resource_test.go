// Copyright © 2024. Citrix Systems, Inc.

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
			TestStorefrontProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFStoreServicePreCheck(t)
			TestSTFAuthenticationServicePreCheck(t)
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
					// Verify pna of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "pna.enable", "true"),
					// Verify enumeration_options of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "enumeration_options.enhanced_enumeration", "false"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "enumeration_options.filter_by_keywords_include.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "enumeration_options.filter_by_keywords_include.1", "AppSet2"),
					// Verify launch_options of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "launch_options.vda_logon_data_provider", "FASLogonDataProvider"),
					// Verify Store Farm Configutations of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.enable_file_type_association", "true"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.communication_timeout", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.connection_timeout", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_failed", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_leasing", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_pending", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.pooled_sockets", "false"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.server_communication_attempts", "5"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.background_healthcheck_polling", "0.0:0:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.advanced_healthcheck", "false"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.cert_revocation_policy", "MustCheck"),
					// Verify roaming_account of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "roaming_account.published", "true"),
					// Verify farms of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.0.farm_name", "Controller1"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.0.farm_type", "XenDesktop"),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_store_service.testSTFStoreService",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "virtual_path",
				ImportStateIdFunc:                    generateImportStateId_STFStoreService,
				ImportStateVerifyIgnore:              []string{"last_updated", "authentication_service_virtual_path", "pna", "enumeration_options", "launch_options", "farm_settings", "roaming_account", "farms"},
			},

			// Update testing for STF Store Service
			{
				Config: BuildSTFStoreServiceResource(t, testSTFStoreServiceResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify friendly_name of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "friendly_name", "Store_Updated"),
					// Verify enumeration_options of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "enumeration_options.enhanced_enumeration", "true"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "enumeration_options.filter_by_keywords_include.#", "1"),
					// Verify launch_options of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "launch_options.vda_logon_data_provider", "UpdatedLogonDataProvider"),
					// Verify Store Farm Configutations of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.enable_file_type_association", "false"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.communication_timeout", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.connection_timeout", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_failed", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_leasing", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.leasing_status_expiry_pending", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.pooled_sockets", "true"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.server_communication_attempts", "4"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.background_healthcheck_polling", "0.0:1:0"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.advanced_healthcheck", "true"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farm_settings.cert_revocation_policy", "NoCheck"),
					// Verify roaming_account of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "roaming_account.published", "false"),
					// Verify farms of STF Store Service
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.0.farm_name", "Controller1"),
					resource.TestCheckResourceAttr("citrix_stf_store_service.testSTFStoreService", "farms.0.farm_type", "XenDesktop"),
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
		authentication_service_virtual_path =  citrix_stf_authentication_service.testSTFAuthenticationService.virtual_path
		enumeration_options = {
			enhanced_enumeration = "false"
			filter_by_keywords_include = ["AppSet1", "AppSet2"]
		}
		pna = {
			enable = "true"
		}
		farm_settings = {
			enable_file_type_association = "true"
			communication_timeout = "0.0:0:0"
			connection_timeout = "0.0:0:0"
			leasing_status_expiry_failed = "0.0:0:0"
			leasing_status_expiry_leasing = "0.0:0:0"
			leasing_status_expiry_pending = "0.0:0:0"
			pooled_sockets = "false"
			server_communication_attempts = "5"
			background_healthcheck_polling = "0.0:0:0"
			advanced_healthcheck = "false"
			cert_revocation_policy = "MustCheck"
		}
		farms = [
    		{
				farm_name = "Controller1"
				farm_type = "XenDesktop"
				servers = ["example-ddc.test.local"]
				port = 80
    		},
			{
				farm_name = "Controller2"
				farm_type = "XenDesktop"
				servers = ["example2-ddc.test.local"]
				port = 80
    		}
  		]
		launch_options = {
        	vda_logon_data_provider = "FASLogonDataProvider"
    	}
		roaming_account = {
			published = "true"
		}
	  }
	`
	testSTFStoreServiceResources_updated = `
	resource "citrix_stf_store_service" "testSTFStoreService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "Store_Updated"
		authentication_service_virtual_path =  citrix_stf_authentication_service.testSTFAuthenticationService.virtual_path
		enumeration_options = {
		enhanced_enumeration = "true"
			filter_by_keywords_include = ["AppSet1"]
		}
		pna = {
			enable = "true"
		}
		farm_settings = {
			enable_file_type_association = "false"
			communication_timeout = "0.0:1:0"
			connection_timeout = "0.0:1:0"
			leasing_status_expiry_failed = "0.0:1:0"
			leasing_status_expiry_leasing = "0.0:1:0"
			leasing_status_expiry_pending = "0.0:1:0"
			pooled_sockets = "true"
			server_communication_attempts = "4"
			background_healthcheck_polling = "0.0:1:0"
			advanced_healthcheck = "true"
			cert_revocation_policy = "NoCheck"
		}
		farms = [
    		{
				farm_name = "Controller1"
				farm_type = "XenDesktop"
				servers = ["example-ddc.test.local"]
				port = 80
    		}
]
		launch_options = {
        	vda_logon_data_provider = "UpdatedLogonDataProvider"
    	}
		roaming_account = {
			published = "false"
		}
	  }
	`
)
