package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_application", &resource.Sweeper{
		Name: "citrix_application",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			appName := os.Getenv("TEST_APP_NAME")
			err := applicationSweeper(ctx, appName, client)
			return err
		},
	})
}

// TestApplicationResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestApplicationResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_APP_NAME"); v == "" {
		t.Fatal("TEST_APP_NAME must be set for acceptance tests")
	}
}

func TestApplicationResource(t *testing.T) {
	name := os.Getenv("TEST_APP_NAME")
	updated_folder_name := fmt.Sprintf("%s-2", os.Getenv("TEST_ADMIN_FOLDER_NAME"))
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestAdminFolderPreCheck(t)
			TestApplicationResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationResource(t, testApplicationResource),
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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
					// Verify the application category path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_category_path", "Main Apps\\Test App"),
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
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_folder_path", updated_folder_name),
					// Verify the application category path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_category_path", ""),
				),
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationResource(t, testApplicationResource_withPriorityModel),
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_folder_path", updated_folder_name),
					// Verify the application category path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_category_path", ""),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_application.testApplication", "delivery_groups_priority.#", "1"),
				),
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildApplicationResource(t, testApplicationResource_withPriorityModel_updated),
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_folder_path", updated_folder_name),
					// Verify the application category path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_category_path", ""),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_application.testApplication", "delivery_groups_priority.#", "1"),
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
	application_category_path = "Main Apps\\Test App"
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
	application_folder_path = citrix_admin_folder.testAdminFolder2.path
}`

	testApplicationResource_withPriorityModel = `
resource "citrix_application" "testApplication" {
	name                = "%s-updated"
	description         = "Application for testing updated"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_arguments  = "update test arguments"
		command_line_executable = "updated_test.exe"
		working_directory       = "test directory"
	}
	delivery_groups_priority = [
		{
			id = citrix_delivery_group.testDeliveryGroup.id
			priority = 0
		}
	]
	application_folder_path = citrix_admin_folder.testAdminFolder2.path
}`
	testApplicationResource_withPriorityModel_updated = `
resource "citrix_application" "testApplication" {
	name                = "%s-updated"
	description         = "Application for testing updated"
	published_name = "TestApplication"
	installed_app_properties = {
		command_line_arguments  = "update test arguments"
		command_line_executable = "updated_test.exe"
		working_directory       = "test directory"
	}
	delivery_groups_priority = [
		{
			id = citrix_delivery_group.testDeliveryGroup.id
			priority = 5
		}
	]
	application_folder_path = citrix_admin_folder.testAdminFolder2.path
}`
)

func BuildApplicationResource(t *testing.T, applicationResource string) string {
	name := os.Getenv("TEST_APP_NAME")
	return fmt.Sprintf(applicationResource, name)
}

func applicationSweeper(ctx context.Context, appName string, client *citrixclient.CitrixDaasClient) error {
	getApplicationRequest := client.ApiClient.ApplicationsAPIsDAAS.ApplicationsGetApplication(ctx, appName)
	application, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.ApplicationDetailResponseModel](getApplicationRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting application: %s", err)
	}
	deleteApplicationRequest := client.ApiClient.ApplicationsAPIsDAAS.ApplicationsDeleteApplication(ctx, application.GetId())
	httpResp, err = citrixclient.AddRequestData(deleteApplicationRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", appName, err)
	}
	return nil
}
