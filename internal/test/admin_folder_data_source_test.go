// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAdminFolderDataSourcePreCheck validates the necessary env variable exist in the testing environment
func TestAdminFolderDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_ADMIN_FOLDER_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_PATH"); v == "" {
		t.Fatal("TEST_ADMIN_FOLDER_DATA_SOURCE_PATH must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_EXPECTED_NAME"); v == "" {
		t.Fatal("TEST_ADMIN_FOLDER_DATA_SOURCE_EXPECTED_NAME must be set for acceptance tests")
	}
}

func TestAdminFolderDataSource(t *testing.T) {
	adminFolderDataSourceId := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_ID")
	adminFolderDataSourcePath := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_PATH")
	expectedAdminFolderDataSourceName := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_EXPECTED_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAdminFolderDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: BuildAdminFolderDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_id", "id", adminFolderDataSourceId),
					// Verify the path of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_id", "path", strings.ReplaceAll(adminFolderDataSourcePath, "\\\\", "\\")),
					// Verify the name attribute of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_id", "name", expectedAdminFolderDataSourceName),
				),
			},
			// Read testing using Name
			{
				Config: BuildAdminFolderDataSourceWithPath(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_path", "id", adminFolderDataSourceId),
					// Verify the path of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_path", "path", strings.ReplaceAll(adminFolderDataSourcePath, "\\\\", "\\")),
					// Verify the name attribute of the admin folder data source
					resource.TestCheckResourceAttr("data.citrix_admin_folder.test_admin_folder_data_source_with_path", "name", expectedAdminFolderDataSourceName),
				),
			},
		},
	})
}

func BuildAdminFolderDataSourceWithId(t *testing.T) string {
	adminFolderId := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_ID")

	return fmt.Sprintf(admin_folder_test_data_source_using_id, adminFolderId)
}

func BuildAdminFolderDataSourceWithPath(t *testing.T) string {
	adminFolderPath := os.Getenv("TEST_ADMIN_FOLDER_DATA_SOURCE_PATH")

	return fmt.Sprintf(admin_folder_test_data_source_using_path, adminFolderPath)
}

var (
	admin_folder_test_data_source_using_id = `
	data "citrix_admin_folder" "test_admin_folder_data_source_with_id" {
		id = "%s"
	}
	`

	admin_folder_test_data_source_using_path = `
	data "citrix_admin_folder" "test_admin_folder_data_source_with_path" {
		path = "%s"
	}
	`
)
