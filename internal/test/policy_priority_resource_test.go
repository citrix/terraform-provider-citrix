// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestPolicyPriorityResource(t *testing.T) {
	policy_set_name := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_NAME")
	policy1Name := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-1"
	policy2Name := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-2"
	policy3Name := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-3"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetV2ResourcePreCheck(t)
			TestPolicyResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildEnabledPolicyResource(t, testPolicy2Resource),
					BuildEnabledPolicyResource(t, testPolicy3Resource),
					testPolicyPriorityResource,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the policy_set_name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_set_name", policy_set_name),
					// Verify the policy_names of the policy resource
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.#", "3"),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.0", policy1Name),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.1", policy2Name),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.2", policy3Name),
				),
			},
			{
				ResourceName:                         "citrix_policy_priority.test_policy_priority",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    generatePolicyPriorityImportStateId,
				ImportStateVerifyIdentifierAttribute: "policy_set_id",
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildEnabledPolicyResource(t, testPolicy2Resource),
					BuildEnabledPolicyResource(t, testPolicy3Resource),
					testPolicyPriorityResourceUpdated,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the policy_set_name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_set_name", policy_set_name),
					// Verify the policy_names of the policy resource
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.#", "3"),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.0", policy3Name),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.1", policy1Name),
					resource.TestCheckResourceAttr("citrix_policy_priority.test_policy_priority", "policy_names.2", policy2Name),
				),
			},
		},
	})
}

func generatePolicyPriorityImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_policy_priority.test_policy_priority"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return rawState["policy_set_id"], nil
}

var (
	testPolicyPriorityResource = `
resource "citrix_policy_priority" "test_policy_priority" {
    policy_set_id    = citrix_policy_set_v2.test_policy_set_v2.id
    policy_priority  = [
        citrix_policy.test_policy1.id,
        citrix_policy.test_policy2.id,
        citrix_policy.test_policy3.id,
    ]
}
`

	testPolicyPriorityResourceUpdated = `
resource "citrix_policy_priority" "test_policy_priority" {
    policy_set_id	 = citrix_policy_set_v2.test_policy_set_v2.id
    policy_priority  = [
	    citrix_policy.test_policy3.id,
        citrix_policy.test_policy1.id,
        citrix_policy.test_policy2.id,
    ]
}
`
)
