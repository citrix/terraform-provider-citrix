// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicyFiltersDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policyFiltersDataSourceTestVariables)
}

func TestPolicyFiltersDataSource(t *testing.T) {
	policyId := os.Getenv("TEST_POLICY_FILTERS_DATA_SOURCE_ID")
	expectedClientIpFilterCount := os.Getenv("TEST_POLICY_FILTERS_DATA_SOURCE_EXPECTED_CLIENT_NAME_FILTER_COUNT")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFiltersDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFiltersDataSource(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_policy_filters.test_policy_filters", "policy_id", policyId),
					resource.TestCheckResourceAttr("data.citrix_policy_filters.test_policy_filters", "client_name_filters.#", expectedClientIpFilterCount),
				),
			},
		},
	})
}

func BuildPolicyFiltersDataSource(t *testing.T) string {
	policyId := os.Getenv("TEST_POLICY_FILTERS_DATA_SOURCE_ID")

	return fmt.Sprintf(policyFiltersTestResource, policyId)
}

var (
	policyFiltersDataSourceTestVariables = []string{
		"TEST_POLICY_FILTERS_DATA_SOURCE_ID",
		"TEST_POLICY_FILTERS_DATA_SOURCE_EXPECTED_CLIENT_NAME_FILTER_COUNT",
	}

	policyFiltersTestResource = `
data "citrix_policy_filters" "test_policy_filters" {
	policy_id = "%s"
}
`
)
