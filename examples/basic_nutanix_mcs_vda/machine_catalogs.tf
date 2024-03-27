resource "citrix_machine_catalog" "example-catalog" {
    name                        = "example-catalog"
    description                 = "description for example catalog"
    provisioning_type 			= "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        identity_type = "ActiveDirectory"
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
        hypervisor = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id
        hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.example-nutanix-rp.id
        nutanix_machine_config = {
            container = "<Container name>"
            master_image = "<Image name>"
            cpu_count = 2
            memory_mb = 4096
            cores_per_cpu_count = 2
        }
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
    }
}