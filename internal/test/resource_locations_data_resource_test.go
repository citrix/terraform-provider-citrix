// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceLocationDataResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_RESOURCE_LOCATION_NAME"); v == "" {
		t.Fatal("TEST_RESOURCE_LOCATION_DATASOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_RESOURCE_LOCATION_ID"); v == "" {
		t.Fatal("TEST_RESOURCE_LOCATION_ID must be set for acceptance tests")
	}
}

func TestResourceLocationDataResource(t *testing.T) {
	name := os.Getenv("TEST_RESOURCE_LOCATION_NAME")
	id := os.Getenv("TEST_RESOURCE_LOCATION_DATASOURCE_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestResourceLocationDataResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildResourceLocationDataResource(t, resourceLocationTestDataResource, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the resource location data source
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "id", id),
				),
			},
		},
	})
}

var (
	resourceLocationTestDataResource = `
resource "citrix_cloud_resource_location" "test_resource_location" {
	name = "%s"
}
`
)

func BuildResourceLocationDataResource(t *testing.T, resourceLocation string, resourceLocationName string) string {
	tfbody := fmt.Sprintf(resourceLocation, resourceLocationName)
	return tfbody
}
