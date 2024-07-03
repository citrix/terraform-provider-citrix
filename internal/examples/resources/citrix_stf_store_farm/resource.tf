resource "citrix_stf_store_farm" "example-stf-store-farm" {
	store_virtual_path      = citrix_stf_store_service.example-stf-store-service.virtual_path
    farm_name = "Controller1"
    farm_type = "XenDesktop"
    servers = ["cvad.storefront.com"] 
    port = 88
    zones = ["Primary","Secondary", "Thirds"]
}

