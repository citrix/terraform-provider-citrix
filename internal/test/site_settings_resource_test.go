// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSiteSettingsResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID"); v == "" {
		t.Fatal("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID must be set for acceptance tests")
	}
}

func TestSiteSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSiteSettingsResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: site_settings_test_resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "web_ui_policy_set_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "dns_resolution_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "multiple_remote_pc_assignments", "true"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "trust_requests_sent_to_the_xml_service_port_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "use_vertical_scaling_for_sessions_on_machines", "false"),
				),
			},
			{
				Config: site_settings_test_resource_updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "web_ui_policy_set_enabled", "true"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "dns_resolution_enabled", "true"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "multiple_remote_pc_assignments", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "trust_requests_sent_to_the_xml_service_port_enabled", "true"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "use_vertical_scaling_for_sessions_on_machines", "true"),
				),
			},
			{
				Config: site_settings_test_resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "web_ui_policy_set_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "dns_resolution_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "multiple_remote_pc_assignments", "true"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "trust_requests_sent_to_the_xml_service_port_enabled", "false"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_citrix_site_settings", "use_vertical_scaling_for_sessions_on_machines", "false"),
				),
			},
		},
	})
}

func TestAdditionalOnPremSiteSettingsResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSiteSettingsResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: onprem_only_site_settings_test_resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "console_inactivity_timeout_minutes", "1440"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "supported_authenticators", "Basic"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "allowed_cors_origins_for_iwa.#", "0"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			{
				Config: onprem_only_site_settings_test_resource_updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "console_inactivity_timeout_minutes", "10"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "supported_authenticators", "All"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "allowed_cors_origins_for_iwa.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_site_settings.test_onprem_citrix_site_settings", "allowed_cors_origins_for_iwa.*", "https://example.com"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			{
				Config: onprem_only_site_settings_test_resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "id", os.Getenv("TEST_SITE_SETTINGS_RESOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "console_inactivity_timeout_minutes", "1440"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "supported_authenticators", "Basic"),
					resource.TestCheckResourceAttr("citrix_site_settings.test_onprem_citrix_site_settings", "allowed_cors_origins_for_iwa.#", "0"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
		},
	})
}

var (
	site_settings_test_resource = `
resource "citrix_site_settings" "test_citrix_site_settings" {
    web_ui_policy_set_enabled                           = false
    dns_resolution_enabled                              = false
    multiple_remote_pc_assignments                      = true
    trust_requests_sent_to_the_xml_service_port_enabled = false
	use_vertical_scaling_for_sessions_on_machines       = false
}`

	site_settings_test_resource_updated = `
	resource "citrix_site_settings" "test_citrix_site_settings" {
		web_ui_policy_set_enabled                           = true
		dns_resolution_enabled                              = true
		multiple_remote_pc_assignments                      = false
		trust_requests_sent_to_the_xml_service_port_enabled = true
		use_vertical_scaling_for_sessions_on_machines       = true
}`

	onprem_only_site_settings_test_resource = `
resource "citrix_site_settings" "test_onprem_citrix_site_settings" {
    console_inactivity_timeout_minutes = 1440
    supported_authenticators           = "Basic"
    allowed_cors_origins_for_iwa       = []
}`

	onprem_only_site_settings_test_resource_updated = `
	resource "citrix_site_settings" "test_onprem_citrix_site_settings" {
		console_inactivity_timeout_minutes = 10
		supported_authenticators           = "All"
		allowed_cors_origins_for_iwa       = ["https://example.com"]
}`
)
