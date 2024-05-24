// Copyright © 2023. Citrix Systems, Inc.

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

func TestSTFDeploymentResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	siteId_updated := os.Getenv("TEST_STF_SITE_ID_UPDATED")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "site_id", siteId),
					// Verify host_base_url of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "host_base_url", "https://test.com"),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_deployment.testSTFDeployment",
				ImportState:                          true,
				ImportStateId:                        siteId,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "site_id",
				ImportStateVerifyIgnore:              []string{"last_updated"},
			},

			// Update testing for STF deployment
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources_updated, siteId_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "site_id", siteId_updated),
					// Verify host_base_url of STF deployment
					resource.TestCheckResourceAttr("citrix_stf_deployment.testSTFDeployment", "host_base_url", "https://test-updated.com"),
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
		host_base_url = "https://test.com"
	}
	`
	testSTFDeploymentResources_updated = `
	resource "citrix_stf_deployment" "testSTFDeployment" {
		site_id      = "%s"
		host_base_url = "https://test-updated.com"
	}
	`
)
