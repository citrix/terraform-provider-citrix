// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAdminUserDataSourcePreCheck validates the necessary env variable exist in the testing environment
func TestAdminUserDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ADMIN_USER_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_ADMIN_USER_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ADMIN_USER_DATA_SOURCE_IS_ENABLED"); v == "" {
		t.Fatal("TEST_ADMIN_USER_DATA_SOURCE_IS_ENABLED must be set for acceptance tests")
	}
}

func TestAdminUserDataSource(t *testing.T) {
	adminUserDataSourceId := os.Getenv("TEST_ADMIN_USER_DATA_SOURCE_ID")
	adminUserDataSourceIsEnabled := os.Getenv("TEST_ADMIN_USER_DATA_SOURCE_IS_ENABLED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAdminUserDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: BuildAdminUserDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin user data source
					resource.TestCheckResourceAttr("data.citrix_admin_user.test_admin_user_data_source_with_id", "id", adminUserDataSourceId),
					// Verify the is_enabled attribute of the admin user data source
					resource.TestCheckResourceAttr("data.citrix_admin_role.test_admin_user_data_source_with_id", "is_enabled", adminUserDataSourceIsEnabled),
				),
			},
		},
	})
}

func BuildAdminUserDataSourceWithId(t *testing.T) string {
	adminUserId := os.Getenv("TEST_ADMIN_USER_DATA_SOURCE_ID")

	return fmt.Sprintf(admin_user_test_data_source_using_id, adminUserId)
}

var (
	admin_user_test_data_source_using_id = `
	data "citrix_admin_user" "test_admin_user_data_source_with_id" {
		id = "%s"
	}
	`
)
