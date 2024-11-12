// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicySetDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_POLICY_SET_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_POLICY_SET_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_EXPECTED_POLICY_COUNT"); v == "" {
		t.Fatal("TEST_POLICY_SET_DATA_SOURCE_EXPECTED_POLICY_COUNT must be set for acceptance tests")
	}
}

func TestPolicySetDataSource(t *testing.T) {
	policySetId := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_ID")
	policySetName := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_NAME")
	policySetExpectedPolicyCount := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_EXPECTED_POLICY_COUNT")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildPolicySetDataSourceById(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "id", policySetId),
					// Verify description of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "name", policySetName),
					// Verify type of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "policies.#", policySetExpectedPolicyCount),
				),
			},
			// Reorder and Read testing
			{
				Config: composeTestResourceTf(
					BuildPolicySetDataSourceByName(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "id", policySetId),
					// Verify description of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "name", policySetName),
					// Verify type of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set.test_policy_set", "policies.#", policySetExpectedPolicyCount),
				),
			},
		},
	})
}

func BuildPolicySetDataSourceById(t *testing.T) string {
	policySetId := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_ID")

	return fmt.Sprintf(policy_set_testResource_by_id, policySetId)
}

func BuildPolicySetDataSourceByName(t *testing.T) string {
	policySetName := os.Getenv("TEST_POLICY_SET_DATA_SOURCE_NAME")

	return fmt.Sprintf(policy_set_testResource_by_name, policySetName)
}

var (
	policy_set_testResource_by_id = `
data "citrix_policy_set" "test_policy_set" {
	id = "%s"
}
`

	policy_set_testResource_by_name = `
data "citrix_policy_set" "test_policy_set" {
	name = "%s"
}
`
)
