resource "citrix_stf_xenapp_default_store" "example-stf-xenapp-default-store" {
	store_virtual_path      = citrix_stf_store_service.example-stf-store-service.virtual_path
    store_site_id           = citrix_stf_store_service.example-stf-store-service.site_id
}

