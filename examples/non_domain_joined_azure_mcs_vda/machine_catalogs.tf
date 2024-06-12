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
        identity_type            = "Workgroup" # Workgroup specifies that the machines are not domain-joined
        azure_machine_config     = {
            service_offering     = var.azure_service_offering
            storage_type         = var.azure_storage_type
            azure_master_image = {
                # shared_subscription = var.azure_image_subscription # Uncomment if the image is from a subscription outside of the hypervisor's subscription

                # For Azure master image from managed disk or snapshot
                resource_group       = var.azure_resource_group
                master_image         = var.azure_master_image

                # For Azure image gallery
                # gallery_image = {
                #     gallery    = var.azure_gallery_name
                #     definition = var.azure_gallery_image_definition
                #     version    = var.azure_gallery_image_version
                # }
            }
            use_managed_disks    = true
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}