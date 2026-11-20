resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type           = "MCS"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor = citrix_nutanix_pc_hypervisor.example-nutanix-pc-hypervisor.id
        hypervisor_resource_pool = citrix_nutanix_pc_hypervisor_resource_pool.example-nutanix-pc-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        nutanix_pc_machine_config = {
            cluster              = var.nutanix_pc_cluster
            master_image         = var.nutanix_pc_master_image
            master_image_version = var.nutanix_pc_master_image_version
            cpu_count            = var.nutanix_pc_cpu_count
            cores_per_cpu_count  = var.nutanix_pc_cores_per_cpu_count
            memory_mb            = var.nutanix_pc_memory_size
        }
        network_mapping = [
            {
                network_device = "0"
                network        = var.nutanix_pc_network
            }
        ]
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}
