// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	policy_set_testResource = `
resource "citrix_policy_set" "testPolicySet" {
    name = "%s-1"
    description = "Test policy set description"
    type = "DeliveryGroupPolicies"
    policies = [
        {
            name = "first-test-policy"
            description = "First test policy with priority 0"
            enabled = true
            policy_settings = [
				{
                    name = "VirtualChannelWhiteList"
                    value = jsonencode([
                        "=disabled="
                    ])
                    use_default = false
                },
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
                {
                    name = "AllowFileDownload"
                    enabled = true
                    use_default = false
                }
            ]
            delivery_group_filters = [
                {
                    delivery_group_id = citrix_delivery_group.testDeliveryGroup.id
                    enabled = true
                    allowed = true
                },
            ]
        },
        {
            name = "second-test-policy"
            description = "Second test policy with priority 1"
            enabled = false
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    use_default = true
                },
            ]
        }
    ]
}
`

	policy_set_reordered_testResource = `
resource "citrix_policy_set" "testPolicySet" {
    name = "%s-2"
    description = "Test policy set description"
    type = "DeliveryGroupPolicies"
    policies = [
		{
            name = "second-test-policy"
            description = "Second test policy with priority 0"
            enabled = false
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    use_default = true
                },
            ]
        },
        {
            name = "first-test-policy"
            description = "First test policy with priority 1"
            enabled = true
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
            ]
            delivery_group_filters = [
                {
                    delivery_group_id = citrix_delivery_group.testDeliveryGroup.id
                    enabled = true
                    allowed = true
                },
            ]
        }
    ]
}
`

	policy_set_updated_testResource = `
resource "citrix_policy_set" "testPolicySet" {
    name = "%s-3"
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
            delivery_group_filters = [
                {
                    delivery_group_id = citrix_delivery_group.testDeliveryGroup.id
                    enabled = true
                    allowed = true
                },
            ]
        }
    ]
}
`
)

func TestPolicySetResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_POLICY_SET_NAME"); v == "" {
		t.Fatal("TEST_POLICY_SET_NAME must be set for acceptance tests")
	}
}

func BuildPolicySetResource(t *testing.T, policySet string) string {
	policySetName := os.Getenv("TEST_POLICY_SET_NAME")

	return fmt.Sprintf(policySet, policySetName)
}

func TestPolicySetResource(t *testing.T) {
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestPolicySetResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildPolicySetResource(t, policy_set_testResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "name", os.Getenv("TEST_POLICY_SET_NAME")+"-1"),
					// Verify description of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "description", "Test policy set description"),
					// Verify type of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "type", "DeliveryGroupPolicies"),
					// Verify the number of scopes of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "scopes.#", "0"),
					// Verify the number of policies in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.#", "2"),
					// Verify name of the first policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.name", "first-test-policy"),
					// Verify policy settings of the first policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("citrix_policy_set.testPolicySet", "policies.0.policy_settings.*", map[string]string{
						"name":        "AdvanceWarningPeriod",
						"use_default": "false",
						"value":       "13:00:00",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("citrix_policy_set.testPolicySet", "policies.0.policy_settings.*", map[string]string{
						"name":        "AllowFileDownload",
						"enabled":     "true",
						"use_default": "false",
					}),
					// Verify name of the second policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.1.name", "second-test-policy"),
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.1.policy_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("citrix_policy_set.testPolicySet", "policies.1.policy_settings.*", map[string]string{
						"name":        "AdvanceWarningPeriod",
						"use_default": "true",
					}),
				),
			},
			// Reorder and Read testing
			{
				Config: composeTestResourceTf(
					BuildPolicySetResource(t, policy_set_reordered_testResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "name", os.Getenv("TEST_POLICY_SET_NAME")+"-2"),
					// Verify description of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "description", "Test policy set description"),
					// Verify type of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "type", "DeliveryGroupPolicies"),
					// Verify the number of scopes of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "scopes.#", "0"),
					// Verify the number of policies in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.#", "2"),
					// Verify name of the first policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.name", "second-test-policy"),
					// Verify name of the second policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.1.name", "first-test-policy"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_policy_set.testPolicySet",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildPolicySetResource(t, policy_set_updated_testResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated, "DesktopsAndApps"),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "name", os.Getenv("TEST_POLICY_SET_NAME")+"-3"),
					// Verify description of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "description", "Test policy set description updated"),
					// Verify type of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "type", "DeliveryGroupPolicies"),
					// Verify the number of scopes of the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "scopes.#", "0"),
					// Verify the number of policies in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.#", "1"),
					// Verify name of the second policy in the policy set
					resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.name", "first-test-policy"),
				),
			},
		},
	})
}
