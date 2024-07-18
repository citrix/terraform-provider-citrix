// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceLocationPreCheck(t *testing.T) {
	if name := os.Getenv("TEST_RESOURCE_LOCATION_NAME"); name == "" {
		t.Fatal("TEST_RESOURCE_LOCATION_NAME must be set for acceptance tests")
	}
}

func TestResourceLocationResource(t *testing.T) {
	name := os.Getenv("TEST_RESOURCE_LOCATION_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestResourceLocationPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildResourceLocationResource(t, resourceLocationTestResource, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the resource location
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "name", name),
					// Verify the value of the internal_only flag (Set to false by default)
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "internal_only", "false"),
					// Verify the value of the time_zone attribute (Set to "UTC" by default)
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "time_zone", "GMT Standard Time"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_cloud_resource_location.test_resource_location",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: BuildResourceLocationResource(t, resourceLocationTestResource_updated, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the resource location
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "name", fmt.Sprintf("%s-updated", name)),
					// Verify the value of the internal_only flag (Set to false by default)
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "internal_only", "false"),
					// Verify the value of the time_zone attribute (Set to "UTC" by default)
					resource.TestCheckResourceAttr("citrix_cloud_resource_location.test_resource_location", "time_zone", "Eastern Standard Time"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	resourceLocationTestResource = `
resource "citrix_cloud_resource_location" "test_resource_location" {
	name = "%s"
}
`
	resourceLocationTestResource_updated = `
resource "citrix_cloud_resource_location" "test_resource_location" {
	name = "%s-updated"
	time_zone = "Eastern Standard Time"
}
`
)

func BuildResourceLocationResource(t *testing.T, resourceLocation string, resourceLocationName string) string {
	tfbody := fmt.Sprintf(resourceLocation, resourceLocationName)
	return tfbody
}
