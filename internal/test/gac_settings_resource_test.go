// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGacSettingsPreCheck(t *testing.T) {
	if service_url := os.Getenv("TEST_SETTINGS_CONFIG_SERVICE_URL"); service_url == "" {
		t.Fatal("TEST_SETTINGS_CONFIG_SERVICE_URL must be set for acceptance tests")
	}
	if name := os.Getenv("TEST_SETTINGS_CONFIG_NAME"); name == "" {
		t.Fatal("TEST_SETTINGS_CONFIG_NAME must be set for acceptance tests")
	}
}

func TestGacSettingsResource(t *testing.T) {
	name := os.Getenv("TEST_SETTINGS_CONFIG_NAME")
	service_url := os.Getenv("TEST_SETTINGS_CONFIG_SERVICE_URL")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestGacSettingsPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildGacSettingsResource(t, gacSettingsTestResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the service url
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "service_url", service_url),
					// Verify the name
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "name", name),
					// Verify the description
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "description", "This is a test resource"),
					// Check permissions for Windows
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.category", "ICA Client"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.0.name", "Allow Client Clipboard Redirection"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.0.value_string", "true"),
					// Check permissions for HTML5
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.category", "Virtual Channel"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.0.name", "Clipboard Operations Between VDA And Local Device"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.0.value_string", "true"),
					// Check permissions for ChromeOS
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.0.category", "Virtual Channel"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.0.settings.0.name", "Clipboard Operations Between VDA And Local Device"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.chromeos.0.settings.0.value_string", "true"),
					// Check permissions for Android
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.0.category", "advanced"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.0.settings.0.name", "enable clipboard"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.android.0.settings.0.value_string", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "citrix_gac_settings.test_settings_configuration",
				ImportState:                          true,
				ImportStateId:                        service_url,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "service_url",
			},
			// Update and Read testing
			{
				Config: BuildGacSettingsResource(t, gacSettingsTestResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the service url
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "service_url", service_url),
					// Verify the name
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "name", fmt.Sprintf("%s - Updated", name)),
					// Verify the description
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "description", "Updated description for test resource"),
					// Check permissions for Windows
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.#", "2"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.category", "ICA Client"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.0.name", "Allow Client Clipboard Redirection"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.0.settings.0.value_string", "true"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.category", "Browser"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.#", "2"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.0.name", "delete browsing data on exit"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.0.value_list.#", "2"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.0.value_list.0", "browsing_history"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.0.value_list.1", "download_history"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.1.name", "relaunch notification period"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.windows.1.settings.1.value_string", "3600000"),
					// Check permissions for HTML5
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.category", "Virtual Channel"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.0.name", "Clipboard Operations Between VDA And Local Device"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.html5.0.settings.0.value_string", "true"),
					// Check permissions for iOS
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.0.category", "Audio"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.0.settings.0.name", "audio"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.ios.0.settings.0.value_string", "true"),
					// Check permissions for MacOS
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.category", "ica client"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.settings.0.name", "Reconnect Apps and Desktops"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.settings.0.value_list.#", "2"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.settings.0.value_list.0", "startWorkspace"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.macos.0.settings.0.value_list.1", "refreshApps"),
					// Check permissions for Linux
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.linux.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.linux.0.category", "root"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.linux.0.user_override", "false"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.linux.0.settings.#", "1"),
					resource.TestCheckResourceAttr("citrix_gac_settings.test_settings_configuration", "app_settings.linux.0.settings.0.name", "enable fido2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	gacSettingsTestResource = `
	resource "citrix_gac_settings" "test_settings_configuration" {
		service_url = "%s"
		name = "%s"
		description = "This is a test resource"
		app_settings = {
			windows = [
				{
					category = "ICA Client",
					user_override = false,
					settings = [
						{
							name = "Allow Client Clipboard Redirection",
							value_string = "true"
						}  
					]
				},
			],
			html5 = [
				{
					category = "Virtual Channel",
					user_override = false,
					settings = [
						{
							name = "Clipboard Operations Between VDA And Local Device",
							value_string = "true"
						}  
					]
				}
			],
			chromeos = [
				{
					category = "Virtual Channel",
					user_override = false,
					settings = [
						{
							name = "Clipboard Operations Between VDA And Local Device",
							value_string = "true"
						}  
					]
				}
			],
			android = [
				{
					category = "advanced",
					user_override = false,
					settings = [
						{
							name = "enable clipboard",
							value_string = "true"
						}  
					]
				}
			]
		}
	}
	`
	gacSettingsTestResource_updated = `
	resource "citrix_gac_settings" "test_settings_configuration" {
		service_url = "%s"
		name = "%s - Updated"
		description = "Updated description for test resource"
		app_settings = {
			windows = [
				{
					user_override = false,
					category = "ICA Client",
					settings = [
						{
							name = "Allow Client Clipboard Redirection",
							value_string = "true"
						}
					]
				},
				{
					user_override = false,
					category = "Browser",
					settings = [
						{
							name = "delete browsing data on exit",
							value_list = [
								"browsing_history",
								"download_history"
							]
						},
						{
							name = "relaunch notification period",
							value_string = "3600000"
						}
					]
				}
			],
			html5 = [
				{
					category = "Virtual Channel",
					user_override = false,
					settings = [
						{
							name = "Clipboard Operations Between VDA And Local Device",
							value_string = "true"
						}  
					]
				}
			],
			ios = [
				{
					category = "Audio",
					user_override = false,
					settings = [
						{
							name = "audio",
							value_string = "true"
						}  
					]
				}
			],
			macos = [
				{
					category = "ica client",
					user_override = false,
					settings = [
						{
							name = "Reconnect Apps and Desktops",
							value_list = [
								"startWorkspace",
								"refreshApps"
							]
						}
					]
				}
			],
			linux = [
				{
					category = "root",
					user_override = false,
					settings = [
						{
							name = "enable fido2",
							value_string = "true"
						}
					]
				}
			]
		}
	}
	`
)

func BuildGacSettingsResource(t *testing.T, settings string) string {
	val := fmt.Sprintf(settings, os.Getenv("TEST_SETTINGS_CONFIG_SERVICE_URL"), os.Getenv("TEST_SETTINGS_CONFIG_NAME"))
	return val
}
