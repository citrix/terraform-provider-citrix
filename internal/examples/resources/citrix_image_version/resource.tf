resource "citrix_image_version" "example_azure_image_version" {
    image_definition = citrix_image_definition.example_image_definition.id
	hypervisor = citrix_azure_hypervisor.example_azure_hypervisor.id   
	hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example_azure_resource_pool.id   
	description = "example description"
	azure_image_specs = {
		service_offering = "Standard_B2als_v2"
		storage_type = "StandardSSD_LRS"
		network_mapping = [
			{
				network_device = "0"
				network 	   = "example_subnet"
			}
		]
		resource_group = "example_resource_group"
        // Only one of master_image and gallery_image can be specified
		master_image = "example_master_image"
        # gallery_image = {
        #     gallery    = var.azure_gallery_name
        #     definition = var.azure_gallery_image_definition
        #     version    = var.azure_gallery_image_version
        # }

        // Optional attributes
		license_type = "Windows_Client"
        shared_subscription = var.azure_image_subscription
        machine_profile = {
            machine_profile_resource_group = var.machine_profile_resource_group

            // Only one of machine_profile_vm_name and machine_profile_template_spec properties can be used
            machine_profile_vm_name = var.machine_profile_vm_name
            # machine_profile_template_spec_name = var.machine_profile_template_spec_name
            # machine_profile_template_spec_version = var.machine_profile_template_spec_version
        }
        disk_encryption_set = {
            disk_encryption_set_name           = var.disk_encryption_set_name
            disk_encryption_set_resource_group = var.disk_encryption_set_resource_group
        }
	}
}