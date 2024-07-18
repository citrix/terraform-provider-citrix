// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestSTFWebReceiverServicePreCheck validates the necessary test API keys exist in the testing environment

func TestSTFWebReceiverServicePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH"); v == "" {
		t.Fatal("TEST_STF_WEBRECEIVER_VIRTUAL_PATH must be set for acceptance tests")
	}
}

func TestSTFWebReceiverServiceResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFAuthenticationServicePreCheck(t)
			TestSTFStoreServicePreCheck(t)
			TestSTFWebReceiverServicePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFWebReceiverServiceResource(t, testSTFWebReceiverServiceResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify site_id of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "site_id", siteId),
					// Verify virtual_path of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "virtual_path", virtualPath),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_webreceiver_service.testSTFWebReceiverService",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "virtual_path",
				ImportStateIdFunc:                    generateImportStateId_STFWebReceiverService,
				ImportStateVerifyIgnore:              []string{"last_updated", "store_virtual_path"},
			},

			// Update testing for STF WebReceiver Service
			{
				Config: BuildSTFWebReceiverServiceResource(t, testSTFWebReceiverServiceResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify friendly_name of STF WebReceiver Service
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "friendly_name", "WebReceiver_Updated"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "authentication_methods.#", "2"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "plugin_assistant.enabled", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "application_shortcuts.trusted_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "application_shortcuts.trusted_urls.*", "https://test.trusted.url/"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "application_shortcuts.gateway_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "application_shortcuts.gateway_urls.*", "https://test.gateway.url/"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.attempts", "3"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.timeout", "0.0:5:0"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.loopback", "On"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.loopback_port_using_http", "8081"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.proxy_enabled", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.proxy_port", "8889"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "communication.proxy_process_name", "TestFiddler"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "strict_transport_security.enabled", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "strict_transport_security.policy_duration", "100.0:0:0"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "authentication_manager.login_form_timeout", "8"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.auto_launch_desktop", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.multi_click_timeout", "5"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.enable_apps_folder_view", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.category_view_collapsed", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.move_app_to_uncategorized", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.show_activity_manager", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.show_first_time_use", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.prevent_ica_downloads", "false"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.workspace_control.enabled", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.workspace_control.auto_reconnect_at_logon", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.workspace_control.logoff_action", "Terminate"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.workspace_control.show_reconnect_button", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.workspace_control.show_disconnect_button", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.receiver_configuration.enabled", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.app_shortcuts.enabled", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.app_shortcuts.allow_session_reconnect", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.ui_views.show_apps_view", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.ui_views.show_desktops_view", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.ui_views.default_view", "Apps"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.progressive_web_app.enabled", "true"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "user_interface.progressive_web_app.show_install_prompt", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "resources_service.ica_file_cache_expiry", "64"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "resources_service.persistent_icon_cache_enabled", "true"),

					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "web_receiver_site_style.header_logo_path", "C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFrontReceiverLogo_Home@2x_B07AF017CEE39553.png"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "web_receiver_site_style.logon_logo_path", "C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFront_auth@2x_CB5D9D1BADB08AFF.png"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "web_receiver_site_style.header_background_color", "Very dark desaturated violet"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "web_receiver_site_style.header_foreground_color", "black"),
					resource.TestCheckResourceAttr("citrix_stf_webreceiver_service.testSTFWebReceiverService", "web_receiver_site_style.link_color", "Dark moderate violet"),
				),
			},
		},
	})
}

func BuildSTFWebReceiverServiceResource(t *testing.T, webreceiverService string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_WEBRECEIVER_VIRTUAL_PATH")

	return BuildSTFStoreServiceResource(t, testSTFStoreServiceResources) + fmt.Sprintf(webreceiverService, siteId, virtualPath)

}

func generateImportStateId_STFWebReceiverService(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_webreceiver_service.testSTFWebReceiverService"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["site_id"], rawState["virtual_path"]), nil
}

var (
	testSTFWebReceiverServiceResources = `
	resource "citrix_stf_webreceiver_service" "testSTFWebReceiverService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "WebReceiver"
		store_virtual_path = citrix_stf_store_service.testSTFStoreService.virtual_path
	
	  }
	`
	testSTFWebReceiverServiceResources_updated = `
	resource "citrix_stf_webreceiver_service" "testSTFWebReceiverService" {
		site_id       = "%s"
		virtual_path = "%s"
		friendly_name = "WebReceiver_Updated" 
		store_virtual_path = citrix_stf_store_service.testSTFStoreService.virtual_path
		authentication_methods = [ 
			"ExplicitForms", 
			"CitrixAGBasic"
		  ]
		plugin_assistant = {
			enabled = true
			html5_single_tab_launch = true
		}
		application_shortcuts = {
			prompt_for_untrusted_shortcuts = true
			trusted_urls                   = [ "https://test.trusted.url/" ]
			gateway_urls                   = [ "https://test.gateway.url/" ]
		}
		communication = {
			attempts = 3
			timeout = "0.0:5:0"
			loopback = "On"
			loopback_port_using_http = 8081
			proxy_enabled = true
			proxy_port = 8889
			proxy_process_name = "TestFiddler"
		}
		strict_transport_security = {
			enabled = true
			policy_duration = "100.0:0:0"
		}
		authentication_manager = {
			login_form_timeout = 8
		}
		user_interface = {
			auto_launch_desktop = true
			multi_click_timeout = 5
			enable_apps_folder_view = true
			workspace_control = {
				enabled = true
				auto_reconnect_at_logon = true
				logoff_action = "Terminate"
				show_reconnect_button = true
				show_disconnect_button = true
			}
			receiver_configuration = {
				enabled = true
			}
			app_shortcuts = {
				enabled = true
				allow_session_reconnect = true	
			}
			ui_views = {
				show_apps_view = true
				show_desktops_view = true
				default_view = "Apps"
			}
			category_view_collapsed = true
			move_app_to_uncategorized = true
			progressive_web_app = {
				enabled = true
				show_install_prompt = true
			}
			show_activity_manager = true
			show_first_time_use = true
			prevent_ica_downloads = false
		}
		resources_service = {
        	ica_file_cache_expiry = 64
    	}	
		web_receiver_site_style = {
			header_logo_path = "C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFrontReceiverLogo_Home@2x_B07AF017CEE39553.png"
			logon_logo_path = "C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFront_auth@2x_CB5D9D1BADB08AFF.png"
			header_background_color = "Very dark desaturated violet"
			header_foreground_color = "black"
			link_color = "Dark moderate violet"
		}
	  }
	`
)
