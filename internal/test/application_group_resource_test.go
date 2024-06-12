package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestApplicationGroupResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestApplicationGroupResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_APP_GROUP_NAME"); v == "" {
		t.Fatal("TEST_APP_GROUP_NAME must be set for acceptance tests")
	}
}

func TestApplicationGroupResource(t *testing.T) {
	name := os.Getenv("TEST_APP_GROUP_NAME")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestApplicationFolderPreCheck(t)
			TestApplicationGroupResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationGroupResource(t, testApplicationGroupResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application_group.testApplicationGroup", "name", name),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application_group.testApplicationGroup", "description", "ApplicationGroup for testing"),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_application_group.testApplicationGroup", "delivery_groups.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_application_group.testApplicationGroup",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"delivery_groups", "installed_app_properties"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationGroupResource(t, testApplicationGroupResource_updated),
					BuildApplicationFolderResource(t, testApplicationFolderResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application_group.testApplicationGroup", "name", fmt.Sprintf("%s-updated", name)),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application_group.testApplicationGroup", "description", "ApplicationGroup for testing updated"),
				),
			},
			// Delete testing
		},
	})
}

var (
	testApplicationGroupResource = `
resource "citrix_application_group" "testApplicationGroup" {
	name                = "%s"
	description         = "ApplicationGroup for testing"
	delivery_groups = [citrix_delivery_group.testDeliveryGroup.id]

}`
	testApplicationGroupResource_updated = `
resource "citrix_application_group" "testApplicationGroup" {
	name                = "%s-updated"
	description         = "ApplicationGroup for testing updated"
	delivery_groups = [citrix_delivery_group.testDeliveryGroup.id]

}`
)

func BuildApplicationGroupResource(t *testing.T, applicationResource string) string {
	name := os.Getenv("TEST_APP_GROUP_NAME")
	return fmt.Sprintf(applicationResource, name)
}
