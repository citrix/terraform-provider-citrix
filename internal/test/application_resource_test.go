package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestApplicationResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestApplicationResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_APP_NAME"); v == "" {
		t.Fatal("TEST_APP_NAME must be set for acceptance tests")
	}
}

func TestApplicationResource(t *testing.T) {
	name := os.Getenv("TEST_APP_NAME")
	updated_folder_name := fmt.Sprintf("%s-2", os.Getenv("TEST_APP_FOLDER_NAME"))
	zoneName := os.Getenv("TEST_ZONE_NAME_AZURE")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestApplicationFolderPreCheck(t)
			TestApplicationResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationResource(t, testApplicationResource),
					BuildApplicationFolderResource(t, testApplicationFolderResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneName, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "name", name),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "description", "Application for testing"),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_application.testApplication", "delivery_groups.#", "1"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_executable", "test.exe"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_application.testApplication",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"delivery_groups", "installed_app_properties"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationResource(t, testApplicationResource_updated),
					BuildApplicationFolderResource(t, testApplicationFolderResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneName, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "name", fmt.Sprintf("%s-updated", name)),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "description", "Application for testing updated"),
					// Verify the command line arguments
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_arguments", "update test arguments"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_executable", "updated_test.exe"),
					// Verify the application folder path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_folder_path", fmt.Sprintf("%s\\", updated_folder_name)),
				),
			},
			// Delete testing
		},
	})
}

var (
	testApplicationResource = `
resource "citrix_application" "testApplication" {
	name                = "%s"
	description         = "Application for testing"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_executable = "test.exe"
		working_directory       = "test directory"
	}
	delivery_groups = [citrix_delivery_group.testDeliveryGroup.id]
}`
	testApplicationResource_updated = `
resource "citrix_application" "testApplication" {
	name                = "%s-updated"
	description         = "Application for testing updated"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_arguments  = "update test arguments"
		command_line_executable = "updated_test.exe"
		working_directory       = "test directory"
	}
	delivery_groups = [citrix_delivery_group.testDeliveryGroup.id]
	application_folder_path = citrix_application_folder.testApplicationFolder2.path
}`
)

func BuildApplicationResource(t *testing.T, applicationResource string) string {
	name := os.Getenv("TEST_APP_NAME")
	return fmt.Sprintf(applicationResource, name)
}
