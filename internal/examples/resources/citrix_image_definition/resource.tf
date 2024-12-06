resource "citrix_image_definition" "example_image_definition" {
    name            = "Example Image Definition"
    description     = "Example Image Definition Description"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor      = citrix_azure_hypervisor.example_azure_hypervisor.id
    azure_image_definition = {
        resource_group = "ExampleResourceGroup"
        use_image_gallery = true
        image_gallery_name = "ExampleImageGalleryName"
    }
}
