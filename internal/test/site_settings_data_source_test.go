// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSiteSettingsDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_SITE_ID"); v == "" {
		t.Fatal("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_SITE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_POLICY_SET_ENABLED"); v == "" {
		t.Fatal("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_POLICY_SET_ENABLED must be set for acceptance tests")
	}
}

func TestSiteSettingsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSiteSettingsDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: BuildSiteSettingsDataSource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_site_settings.test_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("data.citrix_site_settings.test_citrix_site_settings", "web_ui_policy_set_enabled", os.Getenv("TEST_SITE_SETTINGS_DATA_SOURCE_EXPECTED_POLICY_SET_ENABLED")),
				),
			},
		},
	})
}

func BuildSiteSettingsDataSource(t *testing.T) string {
	return site_settings_test_data_source
}

var (
	site_settings_test_data_source = `
	data "citrix_site_settings" "test_citrix_site_settings" { }
	`
)
