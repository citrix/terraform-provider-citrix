// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicyFilterDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policyFilterDataSourceTestVariables)
}

func TestAccessControlPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_ACCESS_CONTROL_POLICY_FILTER_DATA_SOURCE_ID")
	expectedCondition := os.Getenv("TEST_ACCESS_CONTROL_POLICY_FILTER_CONDITION")
	expectedGateway := os.Getenv("TEST_ACCESS_CONTROL_POLICY_FILTER_GATEWAY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, accessControlFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_access_control_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_access_control_policy_filter.test_filter", "condition", expectedCondition),
					resource.TestCheckResourceAttr("data.citrix_access_control_policy_filter.test_filter", "gateway", expectedGateway),
				),
			},
		},
	})
}

func TestBranchRepeaterPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_BRANCH_REPEATER_POLICY_FILTER_DATA_SOURCE_ID")
	branchRepeaterAllowed := os.Getenv("TEST_BRANCH_REPEATER_POLICY_FILTER_DATA_SOURCE_ALLOWED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, branchRepeaterFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_branch_repeater_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_branch_repeater_policy_filter.test_filter", "allowed", branchRepeaterAllowed),
				),
			},
		},
	})
}

func TestClientIPPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_CLIENT_IP_POLICY_FILTER_DATA_SOURCE_ID")
	ipAddress := os.Getenv("TEST_CLIENT_IP_POLICY_FILTER_DATA_SOURCE_IP_ADDRESS")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, clientIpFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_client_ip_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_client_ip_policy_filter.test_filter", "ip_address", ipAddress),
				),
			},
		},
	})
}

func TestClientNamePolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_CLIENT_NAME_POLICY_FILTER_DATA_SOURCE_ID")
	clientName := os.Getenv("TEST_CLIENT_NAME_POLICY_FILTER_DATA_SOURCE_CLIENT_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, clientNameFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_client_name_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_client_name_policy_filter.test_filter", "client_name", clientName),
				),
			},
		},
	})
}

func TestClientPlatformPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_CLIENT_PLATFORM_POLICY_FILTER_DATA_SOURCE_ID")
	platform := os.Getenv("TEST_CLIENT_PLATFORM_POLICY_FILTER_DATA_SOURCE_PLATFORM")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, clientPlatformFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_client_platform_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_client_platform_policy_filter.test_filter", "platform", platform),
				),
			},
		},
	})
}

func TestDeliveryGroupPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_DATA_SOURCE_ID")
	deliveryGroupId := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_DATA_SOURCE_DELIVERY_GROUP_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, deliveryGroupFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_delivery_group_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupId),
				),
			},
		},
	})
}

func TestDeliveryGroupTypePolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_DELIVERY_GROUP_TYPE_POLICY_FILTER_DATA_SOURCE_ID")
	deliveryGroupType := os.Getenv("TEST_DELIVERY_GROUP_TYPE_POLICY_FILTER_DATA_SOURCE_DELIVERY_GROUP_TYPE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, deliveryGroupTypeFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_delivery_group_type_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", deliveryGroupType),
				),
			},
		},
	})
}

func TestOUPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_OU_POLICY_FILTER_DATA_SOURCE_ID")
	expectedOU := os.Getenv("TEST_OU_POLICY_FILTER_DATA_SOURCE_OU")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, ouFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_ou_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_ou_policy_filter.test_filter", "ou", expectedOU),
				),
			},
		},
	})
}

func TestTagPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_TAG_POLICY_FILTER_DATA_SOURCE_ID")
	expectedTag := os.Getenv("TEST_TAG_POLICY_FILTER_DATA_SOURCE_TAG_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, tagFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_tag_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_tag_policy_filter.test_filter", "tag", expectedTag),
				),
			},
		},
	})
}

func TestUserPolicyFilterDataSource(t *testing.T) {
	filterId := os.Getenv("TEST_USER_POLICY_FILTER_DATA_SOURCE_ID")
	expectedUser := os.Getenv("TEST_USER_POLICY_FILTER_DATA_SOURCE_USER")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicyFilterDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicyFilterDataSource(t, userFilterTestDataSource, filterId),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_user_policy_filter.test_filter", "id", filterId),
					resource.TestCheckResourceAttr("data.citrix_user_policy_filter.test_filter", "sid", expectedUser),
				),
			},
		},
	})
}

func BuildPolicyFilterDataSource(t *testing.T, policyFilterResource string, policyFilterId string) string {
	return fmt.Sprintf(policyFilterResource, policyFilterId)
}

var (
	policyFilterDataSourceTestVariables = []string{
		"TEST_ACCESS_CONTROL_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_ACCESS_CONTROL_POLICY_FILTER_CONDITION",
		"TEST_ACCESS_CONTROL_POLICY_FILTER_GATEWAY",
		"TEST_BRANCH_REPEATER_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_BRANCH_REPEATER_POLICY_FILTER_DATA_SOURCE_ALLOWED",
		"TEST_CLIENT_IP_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_CLIENT_IP_POLICY_FILTER_DATA_SOURCE_IP_ADDRESS",
		"TEST_CLIENT_NAME_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_CLIENT_NAME_POLICY_FILTER_DATA_SOURCE_CLIENT_NAME",
		"TEST_CLIENT_PLATFORM_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_CLIENT_PLATFORM_POLICY_FILTER_DATA_SOURCE_PLATFORM",
		"TEST_DELIVERY_GROUP_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_DELIVERY_GROUP_POLICY_FILTER_DATA_SOURCE_DELIVERY_GROUP_ID",
		"TEST_DELIVERY_GROUP_TYPE_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_DELIVERY_GROUP_TYPE_POLICY_FILTER_DATA_SOURCE_DELIVERY_GROUP_TYPE",
		"TEST_OU_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_OU_POLICY_FILTER_DATA_SOURCE_OU",
		"TEST_TAG_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_TAG_POLICY_FILTER_DATA_SOURCE_TAG_ID",
		"TEST_USER_POLICY_FILTER_DATA_SOURCE_ID",
		"TEST_USER_POLICY_FILTER_DATA_SOURCE_USER",
	}

	accessControlFilterTestDataSource = `
data "citrix_access_control_policy_filter" "test_filter" {
	id = "%s"
}
`

	branchRepeaterFilterTestDataSource = `
data "citrix_branch_repeater_policy_filter" "test_filter" {
	id = "%s"
}
`

	clientIpFilterTestDataSource = `
data "citrix_client_ip_policy_filter" "test_filter" {
	id = "%s"
}
`

	clientNameFilterTestDataSource = `
data "citrix_client_name_policy_filter" "test_filter" {
	id = "%s"
}
`

	clientPlatformFilterTestDataSource = `
data "citrix_client_platform_policy_filter" "test_filter" {
	id = "%s"
}
`

	deliveryGroupFilterTestDataSource = `
data "citrix_delivery_group_policy_filter" "test_filter" {
	id = "%s"
}
`

	deliveryGroupTypeFilterTestDataSource = `
data "citrix_delivery_group_type_policy_filter" "test_filter" {
	id = "%s"
}
`

	ouFilterTestDataSource = `
data "citrix_ou_policy_filter" "test_filter" {
	id = "%s"
}
`

	tagFilterTestDataSource = `
data "citrix_tag_policy_filter" "test_filter" {
	id = "%s"
}
`

	userFilterTestDataSource = `
data "citrix_user_policy_filter" "test_filter" {
	id = "%s"
}
`
)
