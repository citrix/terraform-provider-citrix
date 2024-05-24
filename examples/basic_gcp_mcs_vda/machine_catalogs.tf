resource "citrix_machine_catalog" "example-catalog" {
    name                        = "example-gcp-catalog"
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "MCS"
    zone                        = citrix_zone.example-zone.id
    minimum_functional_level    = "L7_20"
    provisioning_scheme         = {
        hypervisor = citrix_gcp_hypervisor.example-gcp-hypervisor.id
        hypervisor_resource_pool = citrix_gcp_hypervisor_resource_pool.example-gcp-rp.id
        identity_type      = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        gcp_machine_config = {
            storage_type = "pd-standard"
            machine_profile = "<Machine profile template VM name>"
            master_image = "<Image template VM name>"
            machine_snapshot = "<Image template VM snapshot name>"
        }
        availability_zones = "<project name>:<region>:<availability zone1>,<project name>:<region>:<availability zone2>,..."
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "ctx-pvdr-##"
            naming_scheme_type = "Numeric"
        }
        writeback_cache = {
			wbc_disk_storage_type = "pd-standard"
			persist_wbc = true
			persist_os_disk = true
			writeback_cache_disk_size_gb = 127
		}
    }
}
