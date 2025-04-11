// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicySetV2DataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policySetV2DataSourceTestVariables)
}

func TestPolicySetV2DataSource(t *testing.T) {
	policySetId := os.Getenv("TEST_POLICY_SET_V2_DATA_SOURCE_ID")
	policySetName := os.Getenv("TEST_POLICY_SET_V2_DATA_SOURCE_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetV2DataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2DataSourceById(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set_v2.test_policy_set", "id", policySetId),
					// Verify name of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set_v2.test_policy_set", "name", policySetName),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2DataSourceByName(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set_v2.test_policy_set", "id", policySetId),
					// Verify name of the policy set
					resource.TestCheckResourceAttr("data.citrix_policy_set_v2.test_policy_set", "name", policySetName),
				),
			},
		},
	})
}

func BuildPolicySetV2DataSourceById(t *testing.T) string {
	policySetId := os.Getenv("TEST_POLICY_SET_V2_DATA_SOURCE_ID")

	return fmt.Sprintf(policy_set_v2_testResource_by_id, policySetId)
}

func BuildPolicySetV2DataSourceByName(t *testing.T) string {
	policySetName := os.Getenv("TEST_POLICY_SET_V2_DATA_SOURCE_NAME")

	return fmt.Sprintf(policy_set_v2_testResource_by_name, policySetName)
}

var (
	policySetV2DataSourceTestVariables = []string{
		"TEST_POLICY_SET_V2_DATA_SOURCE_ID",
		"TEST_POLICY_SET_V2_DATA_SOURCE_NAME",
	}

	policy_set_v2_testResource_by_id = `
data "citrix_policy_set_v2" "test_policy_set" {
	id = "%s"
}
`

	policy_set_v2_testResource_by_name = `
data "citrix_policy_set_v2" "test_policy_set" {
	name = "%s"
}
`
)
