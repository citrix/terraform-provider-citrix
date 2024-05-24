resource "citrix_stf_storeservice" "example-stf-store-service" {
	site_id      = "1"
	virtual_path = "/Citrix/Store"
	friendly_name = "Store"
	authentication_service = "${citrix_stf_authentication_service.example-stf-authentication-service.virtual_path}"
	farm_config = {
		farm_name = "Controller"
		farm_type = "XenDesktop"
		servers = ["cvad.storefront.com", "cvad.storefront2.com"] 
  	}
}

// Anonymous Authentication Service
resource "citrix_stf_store_service" "example-stf-store-service" {
	site_id      = "1"
	virtual_path = "/Citrix/Store"
	friedly_name = "Store"
	anonymous = true
}