resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    provisioning_type 			= "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = "<zone Id>"
    provisioning_scheme         = {
        hypervisor = citrix_xenserver_hypervisor.example-xenserver-hypervisor.id
        hypervisor_resource_pool = citrix_xenserver_hypervisor_resource_pool.example-xenserver-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        xenserver_machine_config = {
            master_image_vm = var.xenserver_master_image_vm
            cpu_count = var.xenserver_cpu_count
            memory_mb = var.xenserver_memory_size
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}