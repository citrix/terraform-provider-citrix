resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
	provisioning_type 			= "MCS"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor = citrix_gcp_hypervisor.example-gcp-hypervisor.id
        hypervisor_resource_pool = citrix_gcp_hypervisor_resource_pool.example-gcp-rp.id
        identity_type      = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        gcp_machine_config = {
            storage_type = var.gcp_storage_type
            master_image = var.gcp_master_image
        }
        availability_zones = var.gcp_availability_zones
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
        writeback_cache = {
			wbc_disk_storage_type = var.gcp_storage_type
			persist_wbc = true
			persist_os_disk = true
			writeback_cache_disk_size_gb = 127
		}
    }
}
