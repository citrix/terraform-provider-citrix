// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicySettingResource(t *testing.T) {
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
					BuildEnabledPolicyResource(t, testPolicy2Resource),
					testPolicySettingResource,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "name", "AdvanceWarningPeriod"),
					// Verify the use_default of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "use_default", "false"),
					// Verify the value of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "value", "13:00:00"),
				),
			},
			{
				ResourceName:      "citrix_policy_setting.test_policy_setting",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildEnabledPolicyResource(t, testPolicy2Resource),
					testPolicySettingResource_Updated,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "name", "AdvanceWarningPeriod"),
					// Verify the use_default of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "use_default", "false"),
					// Verify the value of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "value", "16:00:00"),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySetV2Resource(t),
					BuildEnabledPolicyResource(t, testPolicy1Resource),
					BuildEnabledPolicyResource(t, testPolicy2Resource),
					testPolicySettingResource_UseDefault,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "name", "AdvanceWarningPeriod"),
					// Verify the use_default of the policy setting
					resource.TestCheckResourceAttr("citrix_policy_setting.test_policy_setting", "use_default", "true"),
				),
			},
		},
	})
}

var (
	testPolicySettingResource = `
resource "citrix_policy_setting" "test_policy_setting" {
    policy_id   = citrix_policy.test_policy1.id
    name        = "AdvanceWarningPeriod"
    use_default = false
    value       = "13:00:00"
}
`

	testPolicySettingResource_Updated = `
resource "citrix_policy_setting" "test_policy_setting" {
    policy_id   = citrix_policy.test_policy2.id
    name        = "AdvanceWarningPeriod"
    use_default = false
    value       = "16:00:00"
}
`

	testPolicySettingResource_UseDefault = `
resource "citrix_policy_setting" "test_policy_setting" {
    policy_id   = citrix_policy.test_policy2.id
    name        = "AdvanceWarningPeriod"
    use_default = true
}
`
)
