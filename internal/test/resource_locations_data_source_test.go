// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceLocationDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_RESOURCE_LOCATION_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_RESOURCE_LOCATION_DATASOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_RESOURCE_LOCATION_ID"); v == "" {
		t.Fatal("TEST_RESOURCE_LOCATION_ID must be set for acceptance tests")
	}
}

func TestResourceLocationDataSource(t *testing.T) {
	name := os.Getenv("TEST_RESOURCE_LOCATION_DATASOURCE_NAME")
	id := os.Getenv("TEST_RESOURCE_LOCATION_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestResourceLocationDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildResourceLocationDataSource(t, resourceLocationTestDataSource, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the resource location data source
					resource.TestCheckResourceAttr("data.citrix_cloud_resource_location.test_resource_location", "id", id),
				),
			},
		},
	})
}

var (
	resourceLocationTestDataSource = `
data "citrix_cloud_resource_location" "test_resource_location" {
	name = "%s"
}
`
)

func BuildResourceLocationDataSource(t *testing.T, resourceLocation string, resourceLocationName string) string {
	tfbody := fmt.Sprintf(resourceLocation, resourceLocationName)
	return tfbody
}
