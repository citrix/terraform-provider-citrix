// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestZonePreCheck(t *testing.T) {
	if v := os.Getenv("CITRIX_CUSTOMER_ID"); v != "" && v != "CitrixOnPremises" {
		zoneName := os.Getenv("TEST_ZONE_NAME")
		zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")

		if zoneName == "" || zoneDescription == "" {
			t.Fatal("TEST_ZONE_NAME and TEST_ZONE_DESCRIPTION are required when running tests in cloud env")
		}
	}
}

func TestZoneResource(t *testing.T) {

	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	zoneName := os.Getenv("TEST_ZONE_NAME")
	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")
	if zoneName == "" {
		zoneName = "second zone"
		zoneDescription = "description for go test zone"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { TestProviderPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildZoneResource(t, zone_testResource),
				Check:  getAggregateTestFunc(isOnPremises, zoneName, zoneDescription),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "metadata"},
			},
			// Update and Read testing
			{
				Config: BuildZoneResource(t, zone_testResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", fmt.Sprintf("%s-updated", zoneName)),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", fmt.Sprintf("updated %s", zoneDescription)),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "4"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.name", "key4"),
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.value", "value4"),
				),
				SkipFunc: getSkipFunc(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	zone_testResource = `
resource "citrix_daas_zone" "test" {
	name        = "%s"
	description = "%s"
	metadata    = [
		{
			name = "key1",
			value = "value1"
		},
		{
			name = "key2",
			value = "value2"
		},
		{
			name = "key3",
			value = "value3"
		}
	]
}
`

	zone_testResource_updated = `
resource "citrix_daas_zone" "test" {
	name        = "%s-updated"
	description = "updated %s"
	metadata    = [
		{
			name = "key1",
			value = "value1"
		},
		{
			name = "key2",
			value = "value2"
		},
		{
			name = "key3",
			value = "value3"
		},
		{
			name = "key4",
			value = "value4"
		},
	]
}
`
)

func BuildZoneResource(t *testing.T, zone string) string {
	zoneName := os.Getenv("TEST_ZONE_NAME")
	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")

	if zoneName == "" {
		zoneName = "second zone"
		zoneDescription = "description for go test zone"
	}
	return fmt.Sprintf(zone, zoneName, zoneDescription)
}

func getSkipFunc(isOnPremises bool) func() (bool, error) {
	return func() (bool, error) {
		if isOnPremises {
			return false, nil
		}

		return true, nil
	}
}

func getAggregateTestFunc(isOnPremises bool, zoneName string, zoneDescription string) resource.TestCheckFunc {
	if isOnPremises {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", zoneName),
			resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", zoneDescription),
			resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "3"),
			resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.name", "key1"),
			resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.value", "value1"),
		)
	}

	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", zoneName),
		resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", zoneDescription),
	)
}
