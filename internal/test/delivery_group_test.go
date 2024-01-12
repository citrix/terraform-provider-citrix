// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testHypervisorPreCheck validates the necessary env variable exist
// in the testing environment
func TestDeliveryGroupPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_DG_NAME"); v == "" {
		t.Fatal("TEST_DG_NAME must be set for acceptance tests")
	}
}

func TestDeliveryGroupResourceAzureRM(t *testing.T) {
	name := os.Getenv("TEST_DG_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck(t)
			TestHypervisorResourcePoolPreCheck(t)
			TestMachineCatalogPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildDeliveryGroupResource(t, testDeliveryGroupResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "name", name),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "desktops.#", "2"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "total_machines", "1"),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_daas_delivery_group.testDeliveryGroup",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "autoscale_settings", "associated_machine_catalogs"},
			},

			// Update name, description and add machine testing
			{
				Config: BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "name", fmt.Sprintf("%s-updated", name)),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing updated"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "desktops.#", "1"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "total_machines", "2"),
				),
			},

			// Remove machine testing
			{
				Config: BuildDeliveryGroupResource(t, testDeliveryGroupResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_daas_delivery_group.testDeliveryGroup", "total_machines", "1"),
				),
			},
		},
	})
}

var (
	testDeliveryGroupResources = `
resource "citrix_daas_delivery_group" "testDeliveryGroup" {
    name        = "%s"
    description = "Delivery Group for testing"
	associated_machine_catalogs = [
		{
			machine_catalog = citrix_daas_machine_catalog.testMachineCatalog.id
			machine_count = 1
		}
	]
	desktops = [
		{
            published_name = "desktop-1"
            enabled = true
            enable_session_roaming = true
        },
		{
            published_name = "desktop-2"
            enabled = true
            enable_session_roaming = true
        }
	]
	autoscale_settings = {
		autoscale_enabled = true
		power_time_schemes = [
        	{
        	    "days_of_week" = [
        	        "Monday",
        	        "Tuesday",
        	        "Wednesday",
        	        "Thursday",
        	        "Friday"
        	    ]
        	    "name" = "weekdays test"
        	    "display_name" = "weekdays schedule"
        	    "peak_time_ranges" = [
        	        "09:00-17:00"
        	    ]
        	    "pool_size_schedules": [
        	        {
        	            "time_range": "00:00-00:00",
        	            "pool_size": 1
        	        }
        	    ],
        	    "pool_using_percentage": false
        	},
    	]	
	}
	
}
`
	testDeliveryGroupResources_updated = `
resource "citrix_daas_delivery_group" "testDeliveryGroup" {
    name        = "%s-updated"
    description = "Delivery Group for testing updated"
	associated_machine_catalogs = [
		{
			machine_catalog = citrix_daas_machine_catalog.testMachineCatalog.id
			machine_count = 2
		}
	]
	desktops = [
		{
            published_name = "desktop-1"
            enabled = true
            enable_session_roaming = true
        }
	]
	autoscale_settings = {
		autoscale_enabled = true
		power_time_schemes = [
        	{
        	    "days_of_week" = [
        	        "Monday",
        	        "Tuesday",
        	        "Wednesday",
        	        "Thursday",
        	        "Friday"
        	    ]
        	    "name" = "weekdays test"
        	    "display_name" = "weekdays schedule"
        	    "peak_time_ranges" = [
        	        "09:00-17:00"
        	    ]
        	    "pool_size_schedules": [
        	        {
        	            "time_range": "00:00-00:00",
        	            "pool_size": 1
        	        }
        	    ],
        	    "pool_using_percentage": false
        	},
    	]	
	}
	
}

`
)

func BuildDeliveryGroupResource(t *testing.T, deliveryGroup string) string {
	name := os.Getenv("TEST_DG_NAME")

	return BuildMachineCatalogResource(t, machinecatalog_testResources_updated) + fmt.Sprintf(deliveryGroup, name)
}
