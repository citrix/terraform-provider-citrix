// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_delivery_group", &resource.Sweeper{
		Name: "citrix_delivery_group",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			deliveryGroupName := os.Getenv("TEST_DG_NAME")
			err := deliveryGroupSweeper(ctx, deliveryGroupName, client)
			return err
		},
	})
}

// testHypervisorPreCheck validates the necessary env variable exist
// in the testing environment
func TestDeliveryGroupPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_DG_NAME"); v == "" {
		t.Fatal("TEST_DG_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_POLICY_SET_WITHOUT_DG_NAME"); v == "" {
		t.Fatal("TEST_POLICY_SET_WITHOUT_DG_NAME must be set for acceptance tests")
	}
}

func TestDeliveryGroupResourceAzureRM(t *testing.T) {
	name := os.Getenv("TEST_DG_NAME")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDesktopIconPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsOnly"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "name", name),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing"),
					// Verify delivery type of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "delivery_type", "DesktopsOnly"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "desktops.#", "2"),
					// Verify number of reboot schedules
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "reboot_schedules.#", "2"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "1"),
					// Verify the policy set id assigned to the delivery group
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_delivery_group.testDeliveryGroup",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "autoscale_settings", "associated_machine_catalogs", "reboot_schedules", "delivery_type"},
			},

			// Update name, description and add machine testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "name", fmt.Sprintf("%s-updated", name)),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing updated"),
					// Verify delivery type of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "delivery_type", "DesktopsAndApps"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "desktops.#", "1"),
					// Verify number of reboot schedules
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "reboot_schedules.#", "1"),
					// Verify number of reboot schedules
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "reboot_schedules.0.ignore_maintenance_mode", "false"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "2"),
					// Verify the policy set id assigned to the delivery group
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
			},

			// Remove machine testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsOnly"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "1"),
					// Verify the policy set id assigned to the delivery group
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
					// Verify delivery type of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "delivery_type", "DesktopsOnly"),
				),
			},
			// Update policy set testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updatedWithPolicySetId, ""),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the policy set id assigned to the delivery group
					resource.TestCheckResourceAttrSet("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
			},
		},
	})
}

var (
	testDeliveryGroupResources = `
resource "citrix_delivery_group" "testDeliveryGroup" {
    name        = "%s"
    description = "Delivery Group for testing"
	minimum_functional_level = "L7_9"
	associated_machine_catalogs = [
		{
			machine_catalog = citrix_machine_catalog.testMachineCatalog.id
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
	reboot_schedules = [
		{
			name = "test_reboot_schedule"
			reboot_schedule_enabled = true
			frequency = "Weekly"
			frequency_factor = 1
			days_in_week = [
				"Monday",
				"Tuesday"
				]
			start_time = "12:12"
			start_date = "2024-05-25"
			reboot_duration_minutes = 0
			ignore_maintenance_mode = true
			natural_reboot_schedule = false
			reboot_notification_to_users = {
				notification_duration_minutes = 5
				notification_message = "test message"
				notification_title = "test title"
			}
		},
		{
			name = "test_2"
			reboot_schedule_enabled = true
			frequency = "Monthly"
			frequency_factor = 2
			week_in_month = "First"
			day_in_month = "Monday"
			start_time = "12:12"
			start_date = "2024-04-21"
			ignore_maintenance_mode = true
			reboot_duration_minutes = 120
			natural_reboot_schedule = false
			reboot_notification_to_users = {
				notification_duration_minutes = 15
				notification_message = "test message"
				notification_title = "test title"
				notification_repeat_every_5_minutes = true
			}
		}
	]
	%s
	%s
}
`
	testDeliveryGroupResources_updated = `
resource "citrix_delivery_group" "testDeliveryGroup" {
    name        = "%s-updated"
    description = "Delivery Group for testing updated"
	minimum_functional_level = "L7_20"
	associated_machine_catalogs = [
		{
			machine_catalog = citrix_machine_catalog.testMachineCatalog.id
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
	reboot_schedules = [
		{
			name = "test_reboot_schedule"
			reboot_schedule_enabled = true
			frequency = "Weekly"
			frequency_factor = 1
			days_in_week = [
				"Monday",
				"Tuesday",
				"Wednesday"
				]
			start_time = "12:12"
			start_date = "2024-05-25"
			reboot_duration_minutes = 0
			ignore_maintenance_mode = false
			natural_reboot_schedule = false
		}
	]
	%s
	%s
}

`

	testDeliveryGroupResources_updatedWithPolicySetId = `
resource "citrix_delivery_group" "testDeliveryGroup" {
    name        = "%s-updated"
    description = "Delivery Group for testing updated"
	associated_machine_catalogs = [
		{
			machine_catalog = citrix_machine_catalog.testMachineCatalog.id
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
	reboot_schedules = [
		{
			name = "test_reboot_schedule"
			reboot_schedule_enabled = true
			frequency = "Weekly"
			frequency_factor = 1
			days_in_week = [
				"Monday",
				"Tuesday",
				"Wednesday"
				]
			start_time = "12:12"
			start_date = "2024-05-25"
			reboot_duration_minutes = 0
			ignore_maintenance_mode = true
			natural_reboot_schedule = false
		}
	]
	policy_set_id = citrix_policy_set.testPolicySetWithoutDG.id
	%s
	%s
}

`

	policy_set_no_delivery_group_testResource = `
resource "citrix_policy_set" "testPolicySetWithoutDG" {
    name = "%s"
    description = "Test policy set description updated"
    type = "DeliveryGroupPolicies"
    policies = [
        {
            name = "first-test-policy"
            description = "First test policy with priority 0"
            enabled = true
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
            ]
        }
    ]
}
`
)

func BuildDeliveryGroupResource(t *testing.T, deliveryGroup string, deliveryType string) string {
	name := os.Getenv("TEST_DG_NAME")
	isOnPremises := true
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	if customerId != "" && customerId != "CitrixOnPremises" {
		isOnPremises = false
	}

	deliveryTypeString := ""
	if deliveryType != "" {
		deliveryTypeString = fmt.Sprintf(`delivery_type = "%s"`, deliveryType)
	}

	if isOnPremises {
		return fmt.Sprintf(deliveryGroup, name, deliveryTypeString, "")
	} else {
		return BuildDesktopIconResource(t, testDesktopIconResource) + fmt.Sprintf(deliveryGroup, name, deliveryTypeString, "default_desktop_icon = citrix_desktop_icon.testDesktopIcon.id")
	}

}

func BuildPolicySetResourceWithoutDeliveryGroup(t *testing.T) string {
	policySetName := os.Getenv("TEST_POLICY_SET_WITHOUT_DG_NAME")

	return fmt.Sprintf(policy_set_no_delivery_group_testResource, policySetName)
}

func deliveryGroupSweeper(ctx context.Context, deliveryGroupName string, client *citrixclient.CitrixDaasClient) error {
	getDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupName)
	deliveryGroup, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting delivery group: %s", err)
	}
	deleteDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDeleteDeliveryGroup(ctx, deliveryGroup.GetId())
	httpResp, err = citrixclient.AddRequestData(deleteDeliveryGroupRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", deliveryGroupName, err)
	}
	return nil
}
