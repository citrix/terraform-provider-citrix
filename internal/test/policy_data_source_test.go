// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicyDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policyDataSourceTestVariables)
}

func TestPolicyDataSource(t *testing.T) {
	policyId := os.Getenv("TEST_POLICY_DATA_SOURCE_ID")
	policyName := os.Getenv("TEST_POLICY_DATA_SOURCE_NAME")
	policySetId := os.Getenv("TEST_POLICY_DATA_SOURCE_POLICY_SET_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyDataSourceById(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "id", policyId),
					// Verify name of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "name", policyName),
					// Verify policy_set_id of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "policy_set_id", policySetId),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicyDataSourceByName(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "id", policyId),
					// Verify name of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "name", policyName),
					// Verify policy_set_id of the policy
					resource.TestCheckResourceAttr("data.citrix_policy.test_policy", "policy_set_id", policySetId),
				),
			},
		},
	})
}

func BuildPolicyDataSourceById(t *testing.T) string {
	policyId := os.Getenv("TEST_POLICY_DATA_SOURCE_ID")

	return fmt.Sprintf(policyTestResourceById, policyId)
}

func BuildPolicyDataSourceByName(t *testing.T) string {
	policyName := os.Getenv("TEST_POLICY_DATA_SOURCE_NAME")
	policySetId := os.Getenv("TEST_POLICY_DATA_SOURCE_POLICY_SET_ID")

	return fmt.Sprintf(policyTestResourceByName, policySetId, policyName)
}

var (
	policyDataSourceTestVariables = []string{
		"TEST_POLICY_DATA_SOURCE_ID",
		"TEST_POLICY_DATA_SOURCE_NAME",
		"TEST_POLICY_DATA_SOURCE_POLICY_SET_ID",
	}

	policyTestResourceById = `
data "citrix_policy" "test_policy" {
	id = "%s"
}
`

	policyTestResourceByName = `
data "citrix_policy" "test_policy" {
	policy_set_id = "%s"
	name = "%s"
}
`
)
