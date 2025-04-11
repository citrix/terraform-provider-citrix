// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeliveryGroupPolicyFilterResourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, deliveryGroupPolicyFilterResourceTestVariables)
}

func TestTagPolicyFilterResourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, tagPolicyFilterResourceTestVariables)
}

func TestAccessControlPolicyFilterResource(t *testing.T) {
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
					BuildAccessControlPolicyFilterResource(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "true"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithoutAccessGateway"),
				),
			},
			{
				ResourceName:      "citrix_access_control_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildAccessControlPolicyFilterResource(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "false"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithoutAccessGateway"),
				),
			},
			{
				ResourceName:      "citrix_access_control_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildAccessControlPolicyFilterResource(t, false, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "true"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithoutAccessGateway"),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildAccessControlPolicyFilterResource(t, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "false"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithoutAccessGateway"),
				),
			},
			// Update the delivery group id
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildAccessControlPolicyFilterResourceUpdated(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "true"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithAccessGateway"),
				),
			},
			{
				ResourceName:      "citrix_access_control_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildAccessControlPolicyFilterResourceUpdated(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "allowed", "false"),
					// Verify the connection_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_access_control_policy_filter.test_filter", "connection_type", "WithAccessGateway"),
				),
			},
		},
	})
}

func TestBranchRepeaterPolicyFilterResource(t *testing.T) {
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
					BuildBranchRepeaterPolicyFilterResource(t, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_branch_repeater_policy_filter.test_filter", "allowed", "true"),
				),
			},
			{
				ResourceName:      "citrix_branch_repeater_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildBranchRepeaterPolicyFilterResource(t, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_branch_repeater_policy_filter.test_filter", "allowed", "false"),
				),
			},
		},
	})
}

func TestClientIpPolicyFilterResource(t *testing.T) {
	clientIpAddress := "192.200.0.1"
	clientIpAddressUpdated := "192.200.0.2"
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
					BuildClientIpPolicyFilterResource(t, true, true, clientIpAddress),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "true"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddress),
				),
			},
			{
				ResourceName:      "citrix_client_ip_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientIpPolicyFilterResource(t, true, false, clientIpAddress),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "false"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddress),
				),
			},
			{
				ResourceName:      "citrix_client_ip_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientIpPolicyFilterResource(t, false, true, clientIpAddress),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "true"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddress),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientIpPolicyFilterResource(t, false, false, clientIpAddress),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "false"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddress),
				),
			},
			// Update the client IP address
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientIpPolicyFilterResource(t, true, true, clientIpAddressUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "true"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddressUpdated),
				),
			},
			{
				ResourceName:      "citrix_client_ip_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientIpPolicyFilterResource(t, true, false, clientIpAddressUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "allowed", "false"),
					// Verify the ip_address attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_ip_policy_filter.test_filter", "ip_address", clientIpAddressUpdated),
				),
			},
		},
	})
}

func TestClientNamePolicyFilterResource(t *testing.T) {
	clientName := "test-client-name"
	clientNameUpdated := clientName + "-updated"
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
					BuildClientNamePolicyFilterResource(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "true"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientName),
				),
			},
			{
				ResourceName:      "citrix_client_name_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientNamePolicyFilterResource(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "false"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientName),
				),
			},
			{
				ResourceName:      "citrix_client_name_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientNamePolicyFilterResource(t, false, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "true"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientName),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientNamePolicyFilterResource(t, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "false"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientName),
				),
			},
			// Update the client name
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientNamePolicyFilterResourceUpdated(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "true"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientNameUpdated),
				),
			},
			{
				ResourceName:      "citrix_client_name_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientNamePolicyFilterResourceUpdated(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "allowed", "false"),
					// Verify the client_name attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_name_policy_filter.test_filter", "client_name", clientNameUpdated),
				),
			},
		},
	})
}

func TestClientPlatformPolicyFilterResource(t *testing.T) {
	clientPlatform := "Windows"
	clientPlatformUpdated := "Mac"
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
					BuildClientPlatformPolicyFilterResource(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "true"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatform),
				),
			},
			{
				ResourceName:      "citrix_client_platform_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientPlatformPolicyFilterResource(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "false"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatform),
				),
			},
			{
				ResourceName:      "citrix_client_platform_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientPlatformPolicyFilterResource(t, false, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "true"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatform),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientPlatformPolicyFilterResource(t, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "false"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatform),
				),
			},
			// Update the client name
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientPlatformPolicyFilterResourceUpdated(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "true"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatformUpdated),
				),
			},
			{
				ResourceName:      "citrix_client_platform_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildClientPlatformPolicyFilterResourceUpdated(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "allowed", "false"),
					// Verify the platform attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_client_platform_policy_filter.test_filter", "platform", clientPlatformUpdated),
				),
			},
		},
	})
}

func TestDeliveryGroupPolicyFilterResource(t *testing.T) {
	deliveryGroupId := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID")
	deliveryGroupIdUpdated := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID_UPDATED")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetV2ResourcePreCheck(t)
			TestPolicyResourcePreCheck(t)
			TestDeliveryGroupPolicyFilterResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResource(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupId),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResource(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "false"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupId),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResource(t, false, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupId),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResource(t, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "false"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupId),
				),
			},
			// Update the delivery group id
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResourceUpdated(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupIdUpdated),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupPolicyFilterResourceUpdated(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "allowed", "false"),
					// Verify the delivery_group_id attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_policy_filter.test_filter", "delivery_group_id", deliveryGroupIdUpdated),
				),
			},
		},
	})
}

func TestDeliveryGroupTypePolicyFilterResource(t *testing.T) {
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
					BuildDeliveryGroupTypePolicyFilterResource(t, true, true, "Private"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "Private"),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_type_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, true, false, "Private"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "false"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "Private"),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_type_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, false, true, "Private"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "Private"),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, false, false, "Private"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "false"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "Private"),
				),
			},
			// Update the delivery group type
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, true, true, "PrivateApp"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "PrivateApp"),
				),
			},
			{
				ResourceName:      "citrix_delivery_group_type_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, true, true, "Shared"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "Shared"),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildDeliveryGroupTypePolicyFilterResource(t, true, true, "SharedApp"),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "allowed", "true"),
					// Verify the delivery_group_type attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_delivery_group_type_policy_filter.test_filter", "delivery_group_type", "SharedApp"),
				),
			},
		},
	})
}

func TestOuPolicyFilterResource(t *testing.T) {
	ou := "OU=test,DC=test,DC=local"
	ouUpdated := "OU=testUpdated,DC=test,DC=local"
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
					BuildOuPolicyFilterResource(t, true, true, ou),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "true"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ou),
				),
			},
			{
				ResourceName:      "citrix_ou_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildOuPolicyFilterResource(t, true, false, ou),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "false"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ou),
				),
			},
			{
				ResourceName:      "citrix_ou_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildOuPolicyFilterResource(t, false, true, ou),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "true"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ou),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildOuPolicyFilterResource(t, false, false, ou),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "false"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ou),
				),
			},
			// Update the delivery group id
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildOuPolicyFilterResource(t, true, true, ouUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "true"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ouUpdated),
				),
			},
			{
				ResourceName:      "citrix_ou_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildOuPolicyFilterResource(t, true, false, ouUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "allowed", "false"),
					// Verify the ou attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_ou_policy_filter.test_filter", "ou", ouUpdated),
				),
			},
		},
	})
}

func TestTagPolicyFilterResource(t *testing.T) {
	tagId := os.Getenv("TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID")
	tagIdUpdated := os.Getenv("TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID_UPDATED")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySetV2ResourcePreCheck(t)
			TestPolicyResourcePreCheck(t)
			TestTagDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResource(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "true"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagId),
				),
			},
			{
				ResourceName:      "citrix_tag_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResource(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "false"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagId),
				),
			},
			{
				ResourceName:      "citrix_tag_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResource(t, false, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "true"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagId),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResource(t, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "false"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagId),
				),
			},
			// Update the delivery group id
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResourceUpdated(t, true, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "true"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagIdUpdated),
				),
			},
			{
				ResourceName:      "citrix_tag_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildTagPolicyFilterResourceUpdated(t, true, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "allowed", "false"),
					// Verify the tag attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_tag_policy_filter.test_filter", "tag", tagIdUpdated),
				),
			},
		},
	})
}

func TestUserPolicyFilterResource(t *testing.T) {
	userSid := "S-1-5-21-2403217366-3371035555-1309443259-501"
	userSidUpdated := "S-1-5-21-2403217366-3371035555-1309443259-502"
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
					BuildUserPolicyFilterResourceUpdated(t, true, true, userSid),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "true"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSid),
				),
			},
			{
				ResourceName:      "citrix_user_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildUserPolicyFilterResourceUpdated(t, true, false, userSid),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "false"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSid),
				),
			},
			{
				ResourceName:      "citrix_user_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildUserPolicyFilterResourceUpdated(t, false, true, userSid),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "true"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSid),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildUserPolicyFilterResourceUpdated(t, false, false, userSid),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "false"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "false"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSid),
				),
			},
			// Update the delivery group id
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildUserPolicyFilterResourceUpdated(t, true, true, userSidUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "true"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSidUpdated),
				),
			},
			{
				ResourceName:      "citrix_user_policy_filter.test_filter",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildUserPolicyFilterResourceUpdated(t, true, false, userSidUpdated),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the enabled attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "enabled", "true"),
					// Verify the allowed attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "allowed", "false"),
					// Verify the sid attribute of the policy filter
					resource.TestCheckResourceAttr("citrix_user_policy_filter.test_filter", "sid", userSidUpdated),
				),
			},
		},
	})
}

func BuildAccessControlPolicyFilterResource(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testAccessControlPolicyFilterResource, enabled, allowed)
}

func BuildAccessControlPolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testAccessControlPolicyFilterResourceUpdated, enabled, allowed)
}

func BuildBranchRepeaterPolicyFilterResource(t *testing.T, allowed bool) string {
	return fmt.Sprintf(testBranchRepeaterPolicyFilterResource, allowed)
}

func BuildClientIpPolicyFilterResource(t *testing.T, enabled bool, allowed bool, clientIpAddress string) string {
	return fmt.Sprintf(testClientIpPolicyFilterResource, enabled, allowed, clientIpAddress)
}

func BuildClientNamePolicyFilterResource(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testClientNamePolicyFilterResource, enabled, allowed, "test-client-name")
}

func BuildClientNamePolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testClientNamePolicyFilterResource, enabled, allowed, "test-client-name-updated")
}

func BuildClientPlatformPolicyFilterResource(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testClientPlatformPolicyFilterResource, enabled, allowed, "Windows")
}

func BuildClientPlatformPolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool) string {
	return fmt.Sprintf(testClientPlatformPolicyFilterResource, enabled, allowed, "Mac")
}

func BuildDeliveryGroupPolicyFilterResource(t *testing.T, enabled bool, allowed bool) string {
	deliveryGroupId := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID")
	return fmt.Sprintf(testDeliveryGroupPolicyFilterResource, enabled, allowed, deliveryGroupId)
}

func BuildDeliveryGroupPolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool) string {
	deliveryGroupId := os.Getenv("TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID_UPDATED")
	return fmt.Sprintf(testDeliveryGroupPolicyFilterResource, enabled, allowed, deliveryGroupId)
}

func BuildDeliveryGroupTypePolicyFilterResource(t *testing.T, enabled bool, allowed bool, deliveryGroupType string) string {
	return fmt.Sprintf(testDeliveryGroupTypePolicyFilterResource, enabled, allowed, deliveryGroupType)
}

func BuildOuPolicyFilterResource(t *testing.T, enabled bool, allowed bool, ou string) string {
	return fmt.Sprintf(testOuPolicyFilterResource, enabled, allowed, ou)
}

func BuildTagPolicyFilterResource(t *testing.T, enabled bool, allowed bool) string {
	tagId := os.Getenv("TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID")
	return fmt.Sprintf(testTagPolicyFilterResource, enabled, allowed, tagId)
}

func BuildTagPolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool) string {
	tagId := os.Getenv("TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID_UPDATED")
	return fmt.Sprintf(testTagPolicyFilterResource, enabled, allowed, tagId)
}

func BuildUserPolicyFilterResourceUpdated(t *testing.T, enabled bool, allowed bool, userSid string) string {
	return fmt.Sprintf(testUserPolicyFilterResource, enabled, allowed, userSid)
}

var (
	deliveryGroupPolicyFilterResourceTestVariables = []string{
		"TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID",
		"TEST_DELIVERY_GROUP_POLICY_FILTER_RESOURCE_DELIVERY_GROUP_ID_UPDATED",
	}

	tagPolicyFilterResourceTestVariables = []string{
		"TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID",
		"TEST_TAG_POLICY_FILTER_RESOURCE_TAG_ID_UPDATED",
	}

	testAccessControlPolicyFilterResource = `
resource "citrix_access_control_policy_filter" "test_filter" {
	policy_id   	= citrix_policy.test_policy1.id
	enabled    		= %t
    allowed    		= %t
    connection_type = "WithoutAccessGateway"
    condition  		= "*"
    gateway    		= "*"
}
`

	testAccessControlPolicyFilterResourceUpdated = `
resource "citrix_access_control_policy_filter" "test_filter" {
	policy_id   	= citrix_policy.test_policy1.id
	enabled    		= %t
    allowed    		= %t
    connection_type = "WithAccessGateway"
    condition  		= "*"
    gateway    		= "*"
}
`

	testBranchRepeaterPolicyFilterResource = `
resource "citrix_branch_repeater_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    allowed     = %t
}
`

	testClientIpPolicyFilterResource = `
resource "citrix_client_ip_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    enabled     = %t
    allowed     = %t
    ip_address  = "%s"
}
`

	testClientNamePolicyFilterResource = `
resource "citrix_client_name_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    enabled     = %t
    allowed     = %t
    client_name = "%s"
}
`

	testClientPlatformPolicyFilterResource = `
resource "citrix_client_platform_policy_filter" "test_filter" {
	policy_id   = citrix_policy.test_policy1.id
	enabled     = %t
	allowed     = %t
	platform   = "%s"
}
`

	testDeliveryGroupPolicyFilterResource = `
resource "citrix_delivery_group_policy_filter" "test_filter" {
    policy_id    	  = citrix_policy.test_policy1.id
    enabled      	  = %t
    allowed      	  = %t
    delivery_group_id = "%s"
}
`

	testDeliveryGroupTypePolicyFilterResource = `
resource "citrix_delivery_group_type_policy_filter" "test_filter" {
    policy_id    	  	= citrix_policy.test_policy1.id
    enabled      	  	= %t
    allowed      	  	= %t
    delivery_group_type = "%s"
}
`

	testOuPolicyFilterResource = `
resource "citrix_ou_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    enabled     = %t
    allowed     = %t
    ou	 		= "%s"
}
`

	testTagPolicyFilterResource = `
resource "citrix_tag_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    enabled     = %t
    allowed     = %t
    tag 		= "%s"
}
`

	testUserPolicyFilterResource = `
resource "citrix_user_policy_filter" "test_filter" {
    policy_id   = citrix_policy.test_policy1.id
    enabled     = %t
    allowed     = %t
    sid 		= "%s"
}
`
)
