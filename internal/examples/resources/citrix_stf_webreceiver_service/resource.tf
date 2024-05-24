resource "citrix_stf_webreceiver_service" "example-stf-webreceiver-service"{
  	site_id      = "citrix_stf_deployment.testSTFDeployment.site_id"
	virtual_path = "/Citrix/StoreWeb"
	friendly_name = "Receiver"
  	store_service = "${citrix_stf_store_service.example-stf-store-service.virtual_path}"
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
}
