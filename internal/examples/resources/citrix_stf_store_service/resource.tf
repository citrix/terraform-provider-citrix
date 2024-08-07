resource "citrix_stf_store_service" "example-stf-store-service" {
	site_id      = citrix_stf_deployment.example-stf-deployment.site_id
	virtual_path = "/Citrix/Store"
	friendly_name = "Store"
	authentication_service_virtual_path  = "${citrix_stf_authentication_service.example-stf-authentication-service.virtual_path}"
	pna = {
		enable = true
	}
    enumeration_options = {
        enhanced_enumeration = false
        maximum_concurrent_enumerations = 2
        filter_by_keywords_include = ["AppSet1","AppSet2"]
    }
    launch_options = {
        vda_logon_data_provider = "FASLogonDataProvider"
    }
	farms = [
		{
			farm_name = "Controller1"
			farm_type = "XenDesktop"
			servers = ["cvad.storefront.com"] 
			port = 80
			zones = ["Primary","Secondary", "Thirds"]
		},
		{
			farm_name = "Controller2"
			farm_type = "XenDesktop"
			servers = ["cvad.storefront2.com"] 
			port = 443
			zones = ["Primary"]
		}
	]
	farm_settings = {
		enable_file_type_association = true
		communication_timeout = "0.0:0:0"
		connection_timeout = "0.0:0:0"
		leasing_status_expiry_failed = "0.0:0:0"
		leasing_status_expiry_leasing = "0.0:0:0"
		leasing_status_expiry_pending = "0.0:0:0"
		pooled_sockets = false
		server_communication_attempts = 5
		background_healthcheck_polling = "0.0:0:0"
		advanced_healthcheck = false
		cert_revocation_policy = "MustCheck"
    }

	// Add depends_on attribute to ensure the StoreFront Store with Authentication is created after the Authentication Service
  	depends_on = [ citrix_stf_authentication_service.example-stf-authentication-service ]
}

// Anonymous Authentication Service
resource "citrix_stf_store_service" "example-stf-store-service" {
	site_id      = citrix_stf_deployment.example-stf-deployment.site_id
	virtual_path = "/Citrix/Store"
	friendly_name = "Store"
	anonymous = true
}