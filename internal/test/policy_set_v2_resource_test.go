// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicySetV2ResourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policySetV2ResourceTestVariables)
}

func TestPolicySetV2Resource(t *testing.T) {
	policySetName := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_NAME")
	policySetDescription := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_DESCRIPTION")
	policySetNameUpdated := policySetName + "-updated"
	policySetDescriptionUpdated := policySetDescription + " updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetV2ResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildPolicySetV2Resource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy set resource
					resource.TestCheckResourceAttr("citrix_policy_set_v2.test_policy_set_v2", "name", policySetName),
					// Verify the description of the policy set resource
					resource.TestCheckResourceAttr("citrix_policy_set_v2.test_policy_set_v2", "description", policySetDescription),
				),
			},
			{
				ResourceName:      "citrix_policy_set_v2.test_policy_set_v2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: BuildPolicySetV2Resource_Updated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy set resource
					resource.TestCheckResourceAttr("citrix_policy_set_v2.test_policy_set_v2", "name", policySetNameUpdated),
					// Verify the description of the policy set resource
					resource.TestCheckResourceAttr("citrix_policy_set_v2.test_policy_set_v2", "description", policySetDescriptionUpdated),
				),
			},
		},
	})
}

var (
	policySetV2ResourceTestVariables = []string{
		"TEST_POLICY_SET_V2_RESOURCE_NAME",
		"TEST_POLICY_SET_V2_RESOURCE_DESCRIPTION",
	}

	testPolicySetV2Resource = `
resource "citrix_policy_set_v2" "test_policy_set_v2" {
    name        = "%s"
    description = "%s"
}
`
)

func BuildPolicySetV2Resource(t *testing.T) string {
	name := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_NAME")
	description := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_DESCRIPTION")

	return fmt.Sprintf(testPolicySetV2Resource, name, description)
}

func BuildPolicySetV2Resource_Updated(t *testing.T) string {
	name := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_NAME") + "-updated"
	description := os.Getenv("TEST_POLICY_SET_V2_RESOURCE_DESCRIPTION") + " updated"

	return fmt.Sprintf(testPolicySetV2Resource, name, description)
}
