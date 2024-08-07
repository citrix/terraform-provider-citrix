// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestZonePreCheck(t *testing.T) {
	zoneInput := os.Getenv("TEST_ZONE_INPUT")
	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")

	if zoneInput == "" || zoneDescription == "" {
		t.Fatal("TEST_ZONE_INPUT and TEST_ZONE_DESCRIPTION are required when running tests")
	}
}

func TestCloudZoneCreationPreCheck(t *testing.T) {
	if v := os.Getenv("CITRIX_CUSTOMER_ID"); v != "" && v != "CitrixOnPremises" {
		rlName := os.Getenv("TEST_RESOURCE_LOCATION_NAME")
		if rlName == "" {
			t.Fatal("TEST_RESOURCE_LOCATION_NAME is required when running tests in cloud env")
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

	zoneInput := os.Getenv("TEST_ZONE_INPUT")
	rlName := os.Getenv("TEST_RESOURCE_LOCATION_NAME")

	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")
	if zoneInput == "" {
		zoneInput = "second zone"
		zoneDescription = "description for go test zone"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCloudZoneCreationPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:   BuildZoneResource(t, zoneInput, false),
				Check:    getAggregateTestFunc(isOnPremises, zoneInput, zoneDescription),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "metadata"},
				SkipFunc:                skipForCloud(isOnPremises),
			},
			// Update and Read testing
			{
				Config: BuildZoneResource(t, zoneInput, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "name", fmt.Sprintf("%s-updated", zoneInput)),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "description", fmt.Sprintf("updated %s", zoneDescription)),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.#", "4"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.3.name", "key4"),
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.3.value", "value4"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// Create Zone from Resource Location
			{
				Config: composeTestResourceTf(
					zone_testResource_resource_location,
					BuildResourceLocationResource(t, resourceLocationTestResource, rlName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "name", rlName),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "metadata"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	zone_testResource = `
resource "citrix_zone" "test" {
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
resource "citrix_zone" "test" {
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

	zone_testResource_resource_location = `
resource "citrix_zone" "test" {
	resource_location_id = citrix_cloud_resource_location.test_resource_location.id
}
`

	zone_testResource_cloud_base = `
resource "citrix_zone" "test" {
	resource_location_id = "%s"
	description = "%s"
}
`
)

func BuildZoneResource(t *testing.T, zoneInput string, onPremZoneUpdate bool) string {
	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")

	if v := os.Getenv("CITRIX_CUSTOMER_ID"); v != "" && v != "CitrixOnPremises" {
		// Build cloud zone
		return fmt.Sprintf(zone_testResource_cloud_base, zoneInput, zoneDescription)
	}

	if zoneInput == "" {
		zoneInput = "second zone"
		zoneDescription = "description for go test zone"
	}

	if onPremZoneUpdate {
		return fmt.Sprintf(zone_testResource_updated, zoneInput, zoneDescription)
	} else {
		return fmt.Sprintf(zone_testResource, zoneInput, zoneDescription)
	}
}

func getAggregateTestFunc(isOnPremises bool, zoneInput string, zoneDescription string) resource.TestCheckFunc {
	if isOnPremises {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("citrix_zone.test", "name", zoneInput),
			resource.TestCheckResourceAttr("citrix_zone.test", "description", zoneDescription),
			resource.TestCheckResourceAttr("citrix_zone.test", "metadata.#", "3"),
			resource.TestCheckResourceAttr("citrix_zone.test", "metadata.0.name", "key1"),
			resource.TestCheckResourceAttr("citrix_zone.test", "metadata.0.value", "value1"),
		)
	}

	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("citrix_zone.test", "resource_location_id", zoneInput),
		resource.TestCheckResourceAttr("citrix_zone.test", "description", zoneDescription),
	)
}
