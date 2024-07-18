resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    provisioning_type           = "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor = citrix_scvmm_hypervisor.example-scvmm-hypervisor.id
        hypervisor_resource_pool = citrix_scvmm_hypervisor_resource_pool.example-scvmm-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        scvmm_machine_config = {
            master_image = var.scvmm_master_image
            cpu_count = var.scvmm_cpu_count
            memory_mb = var.scvmm_memory_size
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}