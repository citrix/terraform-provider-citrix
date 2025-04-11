// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicyPriorityDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policyPriorityDataSourceTestVariables)
}

func TestPolicyPriorityDataSource(t *testing.T) {
	policySetId := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_ID")
	policySetName := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_NAME")
	firstPolicyName := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_FIRST_POLICY_NAME")
	secondPolicyName := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_SECOND_POLICY_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyPriorityDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyPriorityDataSourceById(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_set_id", policySetId),
					// Verify name of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_set_name", policySetName),
					// Verify first policy of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_names.0", firstPolicyName),
					// Verify policy_set_id of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_names.1", secondPolicyName),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicyPriorityDataSourceByName(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_set_id", policySetId),
					// Verify name of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_set_name", policySetName),
					// Verify first policy of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_names.0", firstPolicyName),
					// Verify policy_set_id of the policy priority
					resource.TestCheckResourceAttr("data.citrix_policy_priority.test_policy_priority", "policy_names.1", secondPolicyName),
				),
			},
		},
	})
}

func BuildPolicyPriorityDataSourceById(t *testing.T) string {
	policySetId := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_ID")

	return fmt.Sprintf(policyPriorityTestResourceById, policySetId)
}

func BuildPolicyPriorityDataSourceByName(t *testing.T) string {
	policySetId := os.Getenv("TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_NAME")

	return fmt.Sprintf(policyPriorityTestResourceByName, policySetId)
}

var (
	policyPriorityDataSourceTestVariables = []string{
		"TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_ID",
		"TEST_POLICY_PRIORITY_DATA_SOURCE_POLICY_SET_NAME",
		"TEST_POLICY_PRIORITY_DATA_SOURCE_FIRST_POLICY_NAME",
		"TEST_POLICY_PRIORITY_DATA_SOURCE_SECOND_POLICY_NAME",
	}

	policyPriorityTestResourceById = `
data "citrix_policy_priority" "test_policy_priority" {
	policy_set_id = "%s"
}
`

	policyPriorityTestResourceByName = `
data "citrix_policy_priority" "test_policy_priority" {
	policy_set_name = "%s"
}
`
)
