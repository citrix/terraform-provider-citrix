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
	resource.AddTestSweepers("citrix_admin_folder", &resource.Sweeper{
		Name: "citrix_admin_folder",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			adminFolderName := os.Getenv("TEST_ADMIN_FOLDER_NAME")
			err := adminFolderSweeper(ctx, adminFolderName, client)
			return err
		},
		Dependencies: []string{"citrix_application"},
	})
}

// TestAdminFolderPreCheck validates the necessary env variable exist
// in the testing environment
func TestAdminFolderPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ADMIN_FOLDER_NAME"); v == "" {
		t.Fatal("TEST_ADMIN_FOLDER_NAME must be set for acceptance tests")
	}
}

func TestAdminFolderResource(t *testing.T) {
	name := os.Getenv("TEST_ADMIN_FOLDER_NAME")

	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_1_updated := fmt.Sprintf("%s-1-updated", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestAdminFolderPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildAdminFolderResource(t, testAdminFolderResource, "ContainsApplications"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_admin_folder.testAdminFolder2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update type testing
			{
				Config: BuildAdminFolderResource(t, testAdminFolderResource, "ContainsApplicationGroups"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplicationGroups"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplicationGroups"),
				),
			},
			// Update name and parent path testing
			{
				Config: BuildAdminFolderResource(t, testAdminFolderResource_nameAndParentPathUpdated1, "ContainsApplications"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1_updated, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// Update name and remove parent path testing
			{
				Config: BuildAdminFolderResource(t, testAdminFolderResource_parentPathRemoved, "ContainsApplications"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			{
				Config: BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsMachineCatalogs"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsMachineCatalogs"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	testAdminFolderResource = `
resource "citrix_admin_folder" "testAdminFolder1" {
	name = "%s"
	type = ["%s"]
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = ["%s"]
	parent_path = citrix_admin_folder.testAdminFolder1.path
}
`
	testAdminFolderResource_nameAndParentPathUpdated1 = `
resource "citrix_admin_folder" "testAdminFolder1" {
	name = "%s-updated"
	type = ["%s"]
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = ["%s"]
	parent_path = citrix_admin_folder.testAdminFolder1.path
}
`

	testAdminFolderResource_parentPathRemoved = `
resource "citrix_admin_folder" "testAdminFolder1" {
	name = "%s-updated"
	type = ["%s"]
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = ["%s"]
}
`

	testAdminFolderResource_twoTypes = `
resource "citrix_admin_folder" "testAdminFolder1" {
	name = "%s-updated"
	type = ["%s","%s"]
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = ["%s","%s"]
}
`
)

func BuildAdminFolderResource(t *testing.T, adminFolder string, adminFolderType string) string {
	name := os.Getenv("TEST_ADMIN_FOLDER_NAME")
	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	return fmt.Sprintf(adminFolder, folder_name_1, adminFolderType, folder_name_2, adminFolderType)
}

func BuildAdminFolderResourceWithTwoTypes(t *testing.T, adminFolder string, adminFolderType1 string, adminFolderType2 string) string {
	name := os.Getenv("TEST_ADMIN_FOLDER_NAME")
	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	return fmt.Sprintf(adminFolder, folder_name_1, adminFolderType1, adminFolderType2, folder_name_2, adminFolderType1, adminFolderType2)
}

func adminFolderSweeper(ctx context.Context, adminFolderName string, client *citrixclient.CitrixDaasClient) error {
	getAdminFolderRequest := client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolder(ctx, adminFolderName)
	adminFolder, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.AdminFolderResponseModel](getAdminFolderRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting admin folder: %s", err)
	}
	deleteAdminFolderRequest := client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersDeleteAdminFolder(ctx, adminFolder.GetId())
	httpResp, err = citrixclient.AddRequestData(deleteAdminFolderRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", adminFolderName, err)
	}
	return nil
}
