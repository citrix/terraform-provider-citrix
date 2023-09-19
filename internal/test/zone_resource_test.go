package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestZonePreCheck(t *testing.T) {
	if v := os.Getenv("CITRIX_CUSTOMER_ID"); v != "" && v != "CitrixOnPremises" {
		zoneId := os.Getenv("TEST_ZONE_ID")
		zoneName := os.Getenv("TEST_ZONE_NAME")

		if zoneId == "" || zoneName == "" {
			t.Fatal("TEST_ZONE_ID and TEST_ZONE_NAME are required when running tests in cloud env")
		}
	}
}

func TestZoneResource(t *testing.T) {

	customerId := os.Getenv("CITRIX_CUSTOMER_ID")

	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env. Skip zone testing
		return
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { TestProviderPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildZoneResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", "second zone"),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", "description for go test zone"),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "3"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.name", "key1"),
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.value", "value1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: zone_testResource_updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", "second zone - updated"),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", "updated description for go test zone"),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "4"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.name", "key4"),
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.value", "value4"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	zone_testResource = `
	resource "citrix_daas_zone" "test" {
		name        = "%s"
		description = "description for go test zone"
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
		name        = "second zone - updated"
		description = "updated description for go test zone"
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

func BuildZoneResource(t *testing.T) string {
	zoneName := os.Getenv("TEST_ZONE_NAME")
	if zoneName == "" {
		zoneName = "second zone"
	}
	return fmt.Sprintf(zone_testResource, zoneName)
}
