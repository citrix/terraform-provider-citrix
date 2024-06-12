resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type 			= "MCS"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id
        hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.example-nutanix-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        nutanix_machine_config = {
            container = var.nutanix_container
            master_image = var.nutanix_master_image
            cpu_count = var.nutanix_cpu_count
            cores_per_cpu_count = var.nutanix_core_per_cpu_count
            memory_mb = var.nutanix_memory_size
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}