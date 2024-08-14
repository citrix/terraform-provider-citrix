package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type", "ContainsApplications"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_admin_folder.testAdminFolder2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: BuildAdminFolderResource(t, testAdminFolderResource_updated, "ContainsDeliveryGroups"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type", "ContainsDeliveryGroups"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type", "ContainsDeliveryGroups"),
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
	type = "%s"
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = "%s"
	parent_path = citrix_admin_folder.testAdminFolder1.path
}
`
	testAdminFolderResource_updated = `
resource "citrix_admin_folder" "testAdminFolder1" {
	name = "%s-updated"
	type = "%s"
}

resource "citrix_admin_folder" "testAdminFolder2" {
	name = "%s"
	type = "%s"
}
`
)

func BuildAdminFolderResource(t *testing.T, adminFolder string, adminFolderType string) string {
	name := os.Getenv("TEST_ADMIN_FOLDER_NAME")
	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	return fmt.Sprintf(adminFolder, folder_name_1, adminFolderType, folder_name_2, adminFolderType)
}
