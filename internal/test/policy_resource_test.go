// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicyResourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policyResourceTestVariables)
}

func TestPolicyResource(t *testing.T) {
	policyName := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-1"
	policyDescription := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION")
	policyNameUpdated := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-updated-1"
	policyDescriptionUpdated := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION") + " updated"
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
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "name", policyName),
					// Verify the description of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "description", policyDescription),
					// Verify the enabled attribute of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "enabled", "true"),
				),
			},
			{
				ResourceName:      "citrix_policy.test_policy1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource_Updated(t, testPolicy1Resource),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "name", policyNameUpdated),
					// Verify the description of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "description", policyDescriptionUpdated),
					// Verify the enabled attribute of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "enabled", "true"),
				),
			},
			{
				ResourceName:      "citrix_policy.test_policy1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildDisabledPolicyResource_Updated(t, testPolicy1Resource),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "name", policyNameUpdated),
					// Verify the description of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "description", policyDescriptionUpdated),
					// Verify the enabled attribute of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "enabled", "false"),
				),
			},
			{
				ResourceName:      "citrix_policy.test_policy1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildDisabledPolicyResource(t, testPolicy1Resource),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "name", policyName),
					// Verify the description of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "description", policyDescription),
					// Verify the enabled attribute of the policy resource
					resource.TestCheckResourceAttr("citrix_policy.test_policy1", "enabled", "false"),
				),
			},
		},
	})
}

var (
	policyResourceTestVariables = []string{
		"TEST_POLICY_RESOURCE_NAME",
		"TEST_POLICY_RESOURCE_DESCRIPTION",
	}

	testPolicy1Resource = `
resource "citrix_policy" "test_policy1" {
	policy_set_id 	= citrix_policy_set_v2.test_policy_set_v2.id
    name        	= "%s-1"
    description 	= "%s"
    enabled     	= %t
}
`

	testPolicy2Resource = `
resource "citrix_policy" "test_policy2" {
	policy_set_id 	= citrix_policy_set_v2.test_policy_set_v2.id
    name        	= "%s-2"
    description 	= "%s"
    enabled     	= %t
}
`

	testPolicy3Resource = `
resource "citrix_policy" "test_policy3" {
	policy_set_id 	= citrix_policy_set_v2.test_policy_set_v2.id
    name        	= "%s-3"
    description 	= "%s"
    enabled     	= %t
}
`
)

func BuildEnabledPolicyResource(t *testing.T, policyResource string) string {
	name := os.Getenv("TEST_POLICY_RESOURCE_NAME")
	description := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION")

	return fmt.Sprintf(policyResource, name, description, true)
}

func BuildEnabledPolicyResource_Updated(t *testing.T, policyResource string) string {
	name := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-updated"
	description := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION") + " updated"

	return fmt.Sprintf(policyResource, name, description, true)
}

func BuildDisabledPolicyResource(t *testing.T, policyResource string) string {
	name := os.Getenv("TEST_POLICY_RESOURCE_NAME")
	description := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION")

	return fmt.Sprintf(policyResource, name, description, false)
}

func BuildDisabledPolicyResource_Updated(t *testing.T, policyResource string) string {
	name := os.Getenv("TEST_POLICY_RESOURCE_NAME") + "-updated"
	description := os.Getenv("TEST_POLICY_RESOURCE_DESCRIPTION") + " updated"

	return fmt.Sprintf(policyResource, name, description, false)
}
