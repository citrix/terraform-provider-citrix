// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestStoreFrontServerDataSourcePreCheck validates the necessary env variable exist in the testing environment
func TestStoreFrontServerDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_STOREFRONT_SERVER_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_STOREFRONT_SERVER_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_EXPECTED_DESCRIPTION"); v == "" {
		t.Fatal("TEST_STOREFRONT_SERVER_DATA_SOURCE_EXPECTED_DESCRIPTION must be set for acceptance tests")
	}
}

func TestStoreFrontServerDataSource(t *testing.T) {
	storeFrontServerId := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_ID")
	storeFrontServerName := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_NAME")
	expectedDescription := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_EXPECTED_DESCRIPTION")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAdminRoleDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: BuildStoreFrontServerDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_id", "id", storeFrontServerId),
					// Verify the name of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_id", "name", storeFrontServerName),
					// Verify the description attribute of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_id", "description", expectedDescription),
				),
			},
			// Read testing using Name
			{
				Config: BuildStoreFrontServerDataSourceWithName(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_name", "id", storeFrontServerId),
					// Verify the name of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_name", "name", storeFrontServerName),
					// Verify the description attribute of the admin role data source
					resource.TestCheckResourceAttr("data.citrix_storefront_server.test_storefront_server_data_source_with_name", "description", expectedDescription),
				),
			},
		},
	})
}

func BuildStoreFrontServerDataSourceWithId(t *testing.T) string {
	adminRoleId := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_ID")

	return fmt.Sprintf(storefront_server_test_data_source_using_id, adminRoleId)
}

func BuildStoreFrontServerDataSourceWithName(t *testing.T) string {
	adminRoleName := os.Getenv("TEST_STOREFRONT_SERVER_DATA_SOURCE_NAME")

	return fmt.Sprintf(storefront_server_test_data_source_using_name, adminRoleName)
}

var (
	storefront_server_test_data_source_using_id = `
	data "citrix_storefront_server" "test_storefront_server_data_source_with_id" {
		id = "%s"
	}
	`

	storefront_server_test_data_source_using_name = `
	data "citrix_storefront_server" "test_storefront_server_data_source_with_name" {
		name = "%s"
	}
	`
)
