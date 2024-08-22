// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSTFRoamingGatewayPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_NAME"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE_UPDATED"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_URL"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_URL must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_URL_UPDATED"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_URL_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_VERSION"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_VERSION must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_ROAMING_GATEWAY_VERSION_UPDATED"); v == "" {
		t.Fatal("TEST_STF_ROAMING_GATEWAY_VERSION_UPDATED must be set for acceptance tests")
	}
}

func TestSTFRoamingServiceResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	name := os.Getenv("TEST_STF_ROAMING_GATEWAY_NAME")
	logonType := os.Getenv("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE")
	logonTypeUpdated := os.Getenv("TEST_STF_ROAMING_GATEWAY_LOGON_TYPE_UPDATED")
	gatewayUrl := os.Getenv("TEST_STF_ROAMING_GATEWAY_URL")
	gatewayUrlUpdated := os.Getenv("TEST_STF_ROAMING_GATEWAY_URL_UPDATED")
	version := os.Getenv("TEST_STF_ROAMING_GATEWAY_VERSION")
	versionUpdated := os.Getenv("TEST_STF_ROAMING_GATEWAY_VERSION_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestStorefrontProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFRoamingGatewayPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing for STF Roaming Gateway resource
			{
				Config: composeTestResourceTf(
					BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId),
					BuildSTFRoamingGatewayResource(t, testSTFRoamingGatewayResource, siteId, name, logonType, gatewayUrl, version),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "name", name),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "logon_type", logonType),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "gateway_url", gatewayUrl),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "version", version),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_roaming_gateway.testSTFRomaingGateway",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateIdFunc:                    generateImportStateId_STFRoamingGatewayResource,
				ImportStateVerifyIgnore:              []string{"last_updated", "callback_url"},
			},

			// Update testing for STF Roaming Gateway resource
			{
				Config: composeTestResourceTf(
					BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId),
					BuildSTFRoamingGatewayResource(t, testSTFRoamingGatewayResource_updated, siteId, name, logonTypeUpdated, gatewayUrlUpdated, versionUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "name", name),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "logon_type", logonTypeUpdated),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "gateway_url", gatewayUrlUpdated),
					resource.TestCheckResourceAttr("citrix_stf_roaming_gateway.testSTFRomaingGateway", "version", versionUpdated),
				),
			},
		},
	})
}

func generateImportStateId_STFRoamingGatewayResource(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_roaming_gateway.testSTFRomaingGateway"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["site_id"], rawState["name"]), nil
}

func BuildSTFRoamingGatewayResource(t *testing.T, gatewayResource string, siteId string, name string, logonType string, gatewayUrl string, version string) string {
	return fmt.Sprintf(gatewayResource, name, logonType, gatewayUrl, version)
}

var (
	testSTFRoamingGatewayResource = `
	resource "citrix_stf_roaming_gateway" "testSTFRomaingGateway" {
		site_id           = citrix_stf_deployment.testSTFDeployment.site_id
		name		      = "%s"
		logon_type        = "%s"
		gateway_url       = "%s"
		version           = "%s"
		subnet_ip_address = "10.0.0.1"
	}
	`
	testSTFRoamingGatewayResource_updated = `
	resource "citrix_stf_roaming_gateway" "testSTFRomaingGateway" {
		site_id           = citrix_stf_deployment.testSTFDeployment.site_id
		name		      = "%s"
		logon_type        = "%s"
		gateway_url       = "%s"
		version           = "%s"
		subnet_ip_address = "10.0.0.1"
	}
	`
)
