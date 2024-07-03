resource "citrix_stf_webreceiver_service" "example-stf-webreceiver-service"{
  	site_id      = citrix_stf_deployment.example-stf-deployment.site_id
	virtual_path = "/Citrix/StoreWeb"
	friendly_name = "Receiver"
  	store_virtual_path = citrix_stf_store_service.example-stf-store-service.virtual_path
	authentication_methods = [ 
      "ExplicitForms", 
	  "CitrixAGBasic"
    ]
	plugin_assistant = {
		enabled = true
		html5_single_tab_launch = true
		upgrade_at_login = true
		html5_enabled = "Fallback"
	}
	application_shortcuts = {
		prompt_for_untrusted_shortcuts = true
		trusted_urls                   = [ "https://example.trusted.url/" ]
		gateway_urls                   = [ "https://example.gateway.url/" ]
	}
	communication = {
		attempts = 1
		timeout = "0.0:3:0"
		loopback = "Off"
		loopback_port_using_http = 80
		proxy_enabled = false
		proxy_port = 8888
		proxy_process_name = "Fiddler"
	}
	strict_transport_security = {
		enabled = false
		policy_duration = "90.0:0:0"
	}
	authentication_manager = {
		login_form_timeout = 5
	}
	user_interface = {
			auto_launch_desktop = true
			multi_click_timeout = 3
			enable_apps_folder_view = true
			workspace_control = {
				enabled = true
				auto_reconnect_at_logon = true
				logoff_action = "Disconnect"
				show_reconnect_button = false
				show_disconnect_button = false
			}
			receiver_configuration = {
				enabled = true
			}
			app_shortcuts = {
				enabled = true
				show_desktop_shortcut = true	
			}
			ui_views = {
				show_apps_view = true
				show_desktops_view = true
				default_view = "Auto"
			}
			category_view_collapsed = false
			move_app_to_uncategorized = true
			progressive_web_app = {
				enabled = false
				show_install_prompt = false
			}
			show_activity_manager = true
			show_first_time_use = true
			prevent_ica_downloads = false
		}
}
