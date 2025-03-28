resource "citrix_machine_catalog" "example-catalog" {
    name                        = var.machine_catalog_name
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
	provisioning_type 			= "MCS"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor               = citrix_azure_hypervisor.example-azure-hypervisor.id
        hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example-azure-rp.id
        identity_type            = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        azure_machine_config     = {
            prepared_image  = {
                image_definition  = citrix_image_definition.example-image-definition.id
                image_version     = citrix_image_version.example-image-version.id
            }
            service_offering     = var.azure_service_offering
            storage_type         = var.azure_storage_type
            use_managed_disks    = true
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}