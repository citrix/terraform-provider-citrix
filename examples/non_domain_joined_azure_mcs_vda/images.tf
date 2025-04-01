resource "citrix_image_definition" "example-image-definition" {
    name            = var.image_definition_name
    description     = "Description for example image definition"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor               = citrix_azure_hypervisor.example-azure-hypervisor.id
    hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example-azure-rp.id
}

resource "citrix_image_version" "example-image-version" {
    image_definition = citrix_image_definition.example-image-definition.id
    hypervisor               = citrix_azure_hypervisor.example-azure-hypervisor.id
    hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example-azure-rp.id
    description = "Description for example image version"

    azure_image_specs = {
        service_offering     = var.azure_service_offering
        storage_type         = var.azure_storage_type
        # shared_subscription = var.azure_image_subscription # Uncomment if the image is from a subscription outside of the hypervisor's subscription

        # Resource Group is required for any type of Azure master image
        resource_group       = var.azure_resource_group

        # For Azure master image from managed disk or snapshot
        master_image         = var.azure_master_image

        # For Azure image gallery
        # gallery_image = {
        #     gallery    = var.azure_gallery_name
        #     definition = var.azure_gallery_image_definition
        #     version    = var.azure_gallery_image_version
        # }
    }
}