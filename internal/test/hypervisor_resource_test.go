// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testHypervisorPreCheck validates the necessary env variable exist
// in the testing environment
func TestHypervisorPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_NAME"); v == "" {
		t.Fatal("TEST_HYPERV_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_AD_ID"); v == "" {
		t.Fatal("TEST_HYPERV_AD_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SUBSCRIPTION_ID"); v == "" {
		t.Fatal("TEST_HYPERV_SUBSCRIPTION_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_APPLICATION_ID"); v == "" {
		t.Fatal("TEST_HYPERV_APPLICATION_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_APPLICATION_SECRET"); v == "" {
		t.Fatal("TEST_HYPERV_APPLICATION_SECRET must be set for acceptance tests")
	}
}

func TestHypervisorResourceAzureRM(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildHypervisorResource(t, hypervisor_testResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_azure_hypervisor.testHypervisor", "name", name),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_daas_azure_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "application_secret"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResource(t, hypervisor_testResources_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_azure_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

var (
	hypervisor_testResources = `
resource "citrix_daas_azure_hypervisor" "testHypervisor" {
    name                = "%s"
    zone                = %s
    active_directory_id = "%s"
    subscription_id     = "%s"
    application_secret  = "%s"
    application_id      = "%s"
}
`

	hypervisor_testResources_updated = `
resource "citrix_daas_azure_hypervisor" "testHypervisor" {
    name                = "%s-updated"
    zone                = %s
    active_directory_id = "%s"
    subscription_id     = "%s"
    application_secret  = "%s"
    application_id      = "%s"
}
`
)

func BuildHypervisorResource(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME")
	tenantId := os.Getenv("TEST_HYPERV_AD_ID")
	subscriptionId := os.Getenv("TEST_HYPERV_SUBSCRIPTION_ID")
	applicationSecret := os.Getenv("TEST_HYPERV_APPLICATION_SECRET")
	applicationId := os.Getenv("TEST_HYPERV_APPLICATION_ID")

	zoneValueForHypervisor := "citrix_daas_zone.test.id"

	return BuildZoneResource(t, zone_testResource) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, tenantId, subscriptionId, applicationSecret, applicationId)
}
