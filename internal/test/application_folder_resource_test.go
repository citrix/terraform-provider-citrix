package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestApplicationFolderPreCheck validates the necessary env variable exist
// in the testing environment
func TestApplicationFolderPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_APP_FOLDER_NAME"); v == "" {
		t.Fatal("TEST_APP_FOLDER_NAME must be set for acceptance tests")
	}
}

func TestApplicationFolderResource(t *testing.T) {
	name := os.Getenv("TEST_APP_FOLDER_NAME")

	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestApplicationFolderPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildApplicationFolderResource(t, testApplicationFolderResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder1", "name", folder_name_1),
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder2", "name", folder_name_2),
					// Verify parent path of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1)),
					// Verify path of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1, folder_name_2)),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_application_folder.testApplicationFolder2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: BuildApplicationFolderResource(t, testApplicationFolderResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder1", "name", fmt.Sprintf("%s-updated", folder_name_1)),
					// Verify parent path of application
					resource.TestCheckResourceAttr("citrix_daas_application_folder.testApplicationFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	testApplicationFolderResource = `
resource "citrix_daas_application_folder" "testApplicationFolder1" {
	name = "%s"
}

resource "citrix_daas_application_folder" "testApplicationFolder2" {
	name = "%s"
	parent_path = citrix_daas_application_folder.testApplicationFolder1.path
}
`
	testApplicationFolderResource_updated = `
resource "citrix_daas_application_folder" "testApplicationFolder1" {
	name = "%s-updated"
}

resource "citrix_daas_application_folder" "testApplicationFolder2" {
	name = "%s"
}
`
)

func BuildApplicationFolderResource(t *testing.T, applicationFolder string) string {
	name := os.Getenv("TEST_APP_FOLDER_NAME")
	folder_name_1 := fmt.Sprintf("%s-1", name)
	folder_name_2 := fmt.Sprintf("%s-2", name)

	return fmt.Sprintf(applicationFolder, folder_name_1, folder_name_2)
}
