// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestResourceAccessPolicyResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestResourceAccessPolicyResourcePreCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	if v := os.Getenv("TEST_RESOURCE_ACCESS_POLICY_NAME"); v == "" {
		t.Fatal("TEST_RESOURCE_ACCESS_POLICY_NAME must be set for acceptance tests")
	}
}

func TestResourceAccessPolicyResource(t *testing.T) {
	policyName := os.Getenv("TEST_RESOURCE_ACCESS_POLICY_NAME")
	policyName_Updated := policyName + "-updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestResourceAccessPolicyResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: BuildResourceAccessPolicyResource(t, resource_access_policy_test_resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "description", "Resource access policy acceptance test"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enabled", "true"),
					resource.TestCheckResourceAttrPair("citrix_resource_access_policy.test_policy", "tag_id", "citrix_tag.test_policy_tag", "id"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "resource_type", "Application"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enable_criteria_for_include_connections", "true"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "include_connections_criteria_type", "MatchAll"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "include_criteria_filters.#", "2"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "include_criteria_filters.0.filter_name", "TestFarm"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "include_criteria_filters.0.filter_value", "TestIncludeFilter"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enable_criteria_for_exclude_connections", "false"),
					resource.TestCheckNoResourceAttr("citrix_resource_access_policy.test_policy", "exclude_criteria_filters"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "metadata.#", "1"),
				),
			},
			// Import test
			{
				ResourceName:      "citrix_resource_access_policy.test_policy",
				ImportState:       true,
				ImportStateVerify: true,
				// Only metadata already in state is refreshed, so none is imported
				ImportStateVerifyIgnore: []string{"metadata"},
			},
			// Update and Read test
			{
				Config: BuildResourceAccessPolicyResource(t, resource_access_policy_test_resource_updated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("citrix_resource_access_policy.test_policy", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "name", policyName_Updated),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "description", ""),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enabled", "false"),
					resource.TestCheckResourceAttrPair("citrix_resource_access_policy.test_policy", "tag_id", "citrix_tag.test_policy_tag", "id"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "resource_type", "Desktop"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enable_criteria_for_include_connections", "false"),
					resource.TestCheckNoResourceAttr("citrix_resource_access_policy.test_policy", "include_connections_criteria_type"),
					resource.TestCheckNoResourceAttr("citrix_resource_access_policy.test_policy", "include_criteria_filters"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "enable_criteria_for_exclude_connections", "true"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "exclude_criteria_filters.#", "1"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "exclude_criteria_filters.0.filter_value", "TestExcludeFilter"),
					resource.TestCheckNoResourceAttr("citrix_resource_access_policy.test_policy", "metadata"),
				),
			},
			// Removing the tag cannot be done by the update API, so the policy is replaced
			{
				Config: BuildResourceAccessPolicyResource(t, resource_access_policy_test_resource_without_tag),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("citrix_resource_access_policy.test_policy", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "name", policyName_Updated),
					resource.TestCheckNoResourceAttr("citrix_resource_access_policy.test_policy", "tag_id"),
					resource.TestCheckResourceAttr("citrix_resource_access_policy.test_policy", "resource_type", "Desktop"),
				),
			},
		},
	})
}

func BuildResourceAccessPolicyResource(t *testing.T, policyResource string) string {
	policyName := os.Getenv("TEST_RESOURCE_ACCESS_POLICY_NAME")

	return fmt.Sprintf(resource_access_policy_test_tag, policyName) + fmt.Sprintf(policyResource, policyName)
}

var (
	resource_access_policy_test_tag = `
	resource "citrix_tag" "test_policy_tag" {
		name = "%s-tag"
	}
	`

	resource_access_policy_test_resource = `
	resource "citrix_resource_access_policy" "test_policy" {
		name          = "%s"
		description   = "Resource access policy acceptance test"
		tag_id        = citrix_tag.test_policy_tag.id
		resource_type = "Application"

		enable_criteria_for_include_connections = true
		include_connections_criteria_type       = "MatchAll"
		include_criteria_filters = [
			{
				filter_name  = "TestFarm"
				filter_value = "TestIncludeFilter"
			},
			{
				filter_name  = "TestFarm"
				filter_value = "TestSecondIncludeFilter"
			}
		]

		metadata = [
			{
				name  = "key1"
				value = "value1"
			}
		]
	}
	`

	resource_access_policy_test_resource_updated = `
	resource "citrix_resource_access_policy" "test_policy" {
		name          = "%s-updated"
		enabled       = false
		tag_id        = citrix_tag.test_policy_tag.id
		resource_type = "Desktop"

		enable_criteria_for_exclude_connections = true
		exclude_criteria_filters = [
			{
				filter_name  = "TestFarm"
				filter_value = "TestExcludeFilter"
			}
		]
	}
	`

	resource_access_policy_test_resource_without_tag = `
	resource "citrix_resource_access_policy" "test_policy" {
		name          = "%s-updated"
		enabled       = false
		resource_type = "Desktop"

		enable_criteria_for_exclude_connections = true
		exclude_criteria_filters = [
			{
				filter_name  = "TestFarm"
				filter_value = "TestExcludeFilter"
			}
		]
	}
	`
)
