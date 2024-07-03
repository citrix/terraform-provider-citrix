// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestZoneDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestZoneDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_DATASOURCE_ID"); v == "" {
		t.Fatal("TEST_ZONE_DATASOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_ZONE_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_ZONE_DATASOURCE_NAME must be set for acceptance tests")
	}
}

func TestZoneDataSource(t *testing.T) {
	id := os.Getenv("TEST_ZONE_DATASOURCE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZoneDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Name
			{
				Config: BuildZoneDataSource(t, zone_test_data_source_using_name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the zone
					resource.TestCheckResourceAttr("data.citrix_zone.test_zone_by_name", "id", id),
				),
			},
		},
	})
}

func BuildZoneDataSource(t *testing.T, zoneDataSource string) string {
	zoneName := os.Getenv("TEST_ZONE_DATASOURCE_NAME")

	return fmt.Sprintf(zoneDataSource, zoneName)
}

var (
	zone_test_data_source_using_name = `
	data "citrix_zone" "test_zone_by_name" {
		name = "%s"
	}
	`
)
