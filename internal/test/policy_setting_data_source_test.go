// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPolicySettingDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, policySettingDataSourceTestVariables)
}

func TestPolicySettingDataSource(t *testing.T) {
	settingId := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_ID")
	settingName := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_NAME")
	policyId := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_POLICY_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPolicySettingDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildPolicySettingDataSourceById(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "id", settingId),
					// Verify name of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "name", settingName),
					// Verify policy_set_id of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "policy_id", policyId),
				),
			},
			{
				Config: composeTestResourceTf(
					BuildPolicySettingDataSourceByName(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify id of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "id", settingId),
					// Verify name of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "name", settingName),
					// Verify policy_id of the policy setting
					resource.TestCheckResourceAttr("data.citrix_policy_setting.test_policy_setting", "policy_id", policyId),
				),
			},
		},
	})
}

func BuildPolicySettingDataSourceById(t *testing.T) string {
	policyId := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_ID")

	return fmt.Sprintf(policySettingTestResourceById, policyId)
}

func BuildPolicySettingDataSourceByName(t *testing.T) string {
	policySettingName := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_NAME")
	policyId := os.Getenv("TEST_POLICY_SETTING_DATA_SOURCE_POLICY_ID")

	return fmt.Sprintf(policySettingTestResourceByName, policyId, policySettingName)
}

var (
	policySettingDataSourceTestVariables = []string{
		"TEST_POLICY_SETTING_DATA_SOURCE_ID",
		"TEST_POLICY_SETTING_DATA_SOURCE_NAME",
		"TEST_POLICY_SETTING_DATA_SOURCE_POLICY_ID",
	}

	policySettingTestResourceById = `
data "citrix_policy_setting" "test_policy_setting" {
	id = "%s"
}
`

	policySettingTestResourceByName = `
data "citrix_policy_setting" "test_policy_setting" {
	policy_id = "%s"
	name = "%s"
}
`
)
