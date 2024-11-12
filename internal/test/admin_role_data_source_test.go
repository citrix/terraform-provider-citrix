// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAdminRoleDataSourcePreCheck validates the necessary env variable exist in the testing environment
func TestAdminRoleDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_ADMIN_ROLE_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_ADMIN_ROLE_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_EXPECTED_DESCRIPTION"); v == "" {
		t.Fatal("TEST_ADMIN_ROLE_DATA_SOURCE_EXPECTED_DESCRIPTION must be set for acceptance tests")
	}
}

func TestAdminRoleDataSource(t *testing.T) {
	adminRoleDataSourceId := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_ID")
	adminRoleDataSourceName := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_NAME")
	expectedDescription := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_EXPECTED_DESCRIPTION")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAdminRoleDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: BuildAdminRoleDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_id", "id", adminRoleDataSourceId),
					// Verify the name of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_id", "name", adminRoleDataSourceName),
					// Verify the description attribute of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_id", "description", expectedDescription),
				),
			},
			// Read testing using Name
			{
				Config: BuildAdminRoleDataSourceWithName(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_name", "id", adminRoleDataSourceId),
					// Verify the name of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_name", "name", adminRoleDataSourceName),
					// Verify the description attribute of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_role_data_source_with_name", "description", expectedDescription),
				),
			},
		},
	})
}

func BuildAdminRoleDataSourceWithId(t *testing.T) string {
	adminRoleId := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_ID")

	return fmt.Sprintf(admin_role_test_data_source_using_id, adminRoleId)
}

func BuildAdminRoleDataSourceWithName(t *testing.T) string {
	adminRoleName := os.Getenv("TEST_ADMIN_ROLE_DATA_SOURCE_NAME")

	return fmt.Sprintf(admin_role_test_data_source_using_name, adminRoleName)
}

var (
	admin_role_test_data_source_using_id = `
	data "citrix_admin_role" "test_admin_role_data_source_with_id" {
		id = "%s"
	}
	`

	admin_role_test_data_source_using_name = `
	data "citrix_admin_role" "test_admin_role_data_source_with_name" {
		name = "%s"
	}
	`
)
