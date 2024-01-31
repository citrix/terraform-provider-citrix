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
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
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
				Config: BuildApplicationResource(t, testApplicationResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "name", name),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "description", "Application for testing"),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "delivery_groups.#", "1"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "installed_app_properties.command_line_executable", "test.exe"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_application.testApplication",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"delivery_groups", "installed_app_properties"},
			},
			// Update and Read testing
			{
				Config: BuildApplicationResource(t, testApplicationResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "name", fmt.Sprintf("%s-updated", name)),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "description", "Application for testing updated"),
					// Verify the command line arguments
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "installed_app_properties.command_line_arguments", "update test arguments"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "installed_app_properties.command_line_executable", "updated_test.exe"),
					// Verify the application folder path
					resource.TestCheckResourceAttr("citrix_daas_application.testApplication", "application_folder_path", fmt.Sprintf("%s\\", updated_folder_name)),
				),
			},
			// Delete testing
		},
	})
}

var (
	testApplicationResource = `
resource "citrix_daas_application" "testApplication" {
	name                = "%s"
	description         = "Application for testing"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_executable = "test.exe"
		working_directory       = "test directory"
	}
	delivery_groups = [citrix_daas_delivery_group.testDeliveryGroup.id]
}`
	testApplicationResource_updated = `
resource "citrix_daas_application" "testApplication" {
	name                = "%s-updated"
	description         = "Application for testing updated"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_arguments  = "update test arguments"
		command_line_executable = "updated_test.exe"
		working_directory       = "test directory"
	}
	delivery_groups = [citrix_daas_delivery_group.testDeliveryGroup.id]
	application_folder_path = citrix_daas_application_folder.testApplicationFolder2.path
}`
)

func BuildApplicationResource(t *testing.T, applicationResource string) string {
	name := os.Getenv("TEST_APP_NAME")
	return BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated) + BuildApplicationFolderResource(t, testApplicationFolderResource_updated) + fmt.Sprintf(applicationResource, name)
}
