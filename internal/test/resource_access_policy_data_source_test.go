// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceAccessPolicyDataSource(t *testing.T) {
	policyName := os.Getenv("TEST_RESOURCE_ACCESS_POLICY_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestResourceAccessPolicyResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read by name and by id
			{
				Config: BuildResourceAccessPolicyResource(t, resource_access_policy_test_resource) + resource_access_policy_test_data_sources,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.citrix_resource_access_policy.test_policy_by_name", "id", "citrix_resource_access_policy.test_policy", "id"),
					resource.TestCheckResourceAttr("data.citrix_resource_access_policy.test_policy_by_name", "name", policyName),
					resource.TestCheckResourceAttrPair("data.citrix_resource_access_policy.test_policy_by_name", "tag_id", "citrix_tag.test_policy_tag", "id"),
					resource.TestCheckResourceAttrPair("data.citrix_resource_access_policy.test_policy_by_name", "tag_name", "citrix_tag.test_policy_tag", "name"),
					resource.TestCheckResourceAttr("data.citrix_resource_access_policy.test_policy_by_name", "resource_type", "Application"),
					resource.TestCheckResourceAttr("data.citrix_resource_access_policy.test_policy_by_name", "include_connections_criteria_type", "MatchAll"),
					resource.TestCheckResourceAttr("data.citrix_resource_access_policy.test_policy_by_name", "include_criteria_filters.#", "2"),
					resource.TestCheckResourceAttrPair("data.citrix_resource_access_policy.test_policy_by_id", "name", "citrix_resource_access_policy.test_policy", "name"),
				),
			},
		},
	})
}

var resource_access_policy_test_data_sources = `
	data "citrix_resource_access_policy" "test_policy_by_name" {
		name = citrix_resource_access_policy.test_policy.name
	}

	data "citrix_resource_access_policy" "test_policy_by_id" {
		id = citrix_resource_access_policy.test_policy.id
	}
	`
