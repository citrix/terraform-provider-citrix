// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment

func TestSTFDeploymentPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_STF_SITE_ID_UPDATED"); v == "" {
		t.Fatal("TEST_STF_SITE_ID_UPDATED must be set for acceptance tests")
	}

}

func TestSTFDeploymentResourceWithProperties(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	siteId_updated := os.Getenv("TEST_STF_SITE_ID_UPDATED")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestStorefrontProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create Deployment along with Roaming Gateway and Roaming Beacon
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "site_id", siteId),
					// Verify host_base_url of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "host_base_url", "http://test"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.name", "Example Roaming Gateway Name"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.logon_type", "Domain"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.gateway_url", "https://example1.gateway.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.subnet_ip_address", "10.0.0.10"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.internal_address", "https://example.internal.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.0", "https://example.external.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.1", "https://example1.external.url/"),
				),
			},

			// Update testing for Deployment, Roaming Gateway, and Roaming Beacon
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources_updated, siteId_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "site_id", siteId_updated),
					// Verify host_base_url of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "host_base_url", "http://testupdated"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.name", "Example Roaming Gateway Name"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.logon_type", "None"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.gateway_url", "https://example.gateway.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.subnet_ip_address", "10.0.0.1"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.internal_address", "https://example1.internal.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.0", "https://example.external.url/"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "citrix_stf_deployment.testSTFDeployment",
				ImportState:                          true,
				ImportStateId:                        siteId_updated,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "site_id",
				ImportStateVerifyIgnore:              []string{"last_updated", "roaming_beacon.external_addresses", "roaming_gateway"},
			},
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources_OnlyInt, siteId_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "site_id", siteId_updated),
					// Verify host_base_url of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "host_base_url", "http://testupdated"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.name", "Example Roaming Gateway Name"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.logon_type", "None"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.gateway_url", "https://example.gateway.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_gateway.0.subnet_ip_address", "10.0.0.1"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.internal_address", "https://example1.internal.url/"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "roaming_beacon.external_addresses.0", "https://example.external.url/"),
				),
			},
		},
	})

}

func BuildSTFDeploymentResource(t *testing.T, deployment string, siteId string) string {
	return fmt.Sprintf(deployment, siteId)
}

var (
	testSTFDeploymentResources = `
	resource "citrix_stf_deployment" "testSTFDeployment" {
		site_id      = "%s"
		host_base_url = "http://test"
		roaming_gateway = [{
		    name = "Example Roaming Gateway Name"
			logon_type = "Domain"
			gateway_url = "https://example1.gateway.url/"
			subnet_ip_address = "10.0.0.10"
		}]
		roaming_beacon = {
			internal_address = "https://example.internal.url/"
			external_addresses = ["https://example.external.url/", "https://example1.external.url/"]
		}
	}
	`
	testSTFDeploymentResources_updated = `
	resource "citrix_stf_deployment" "testSTFDeployment" {
		site_id      = "%s"
		host_base_url = "http://testupdated"
		roaming_gateway = [{
		    name = "Example Roaming Gateway Name"
			logon_type = "None"
			gateway_url = "https://example.gateway.url/"
			subnet_ip_address = "10.0.0.1"
		}]
		roaming_beacon = {
			internal_address = "https://example1.internal.url/"
			external_addresses = ["https://example.external.url/"]
		}
	}
	`
	testSTFDeploymentResources_OnlyInt = `
	resource "citrix_stf_deployment" "testSTFDeployment" {
		site_id      = "%s"
		host_base_url = "http://testupdated"
		roaming_gateway = [{
		    name = "Example Roaming Gateway Name"
			logon_type = "None"
			gateway_url = "https://example.gateway.url/"
			subnet_ip_address = "10.0.0.1"
		}]
		roaming_beacon = {
			internal_address = "https://example1.internal.url/"
		}
	}
	`
)
