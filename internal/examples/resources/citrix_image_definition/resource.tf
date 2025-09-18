resource "citrix_image_definition" "example_azure_image_definition" {
    name            = "Example Azure Image Definition"
    description     = "Example Azure Image Definition Description"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor      = citrix_azure_hypervisor.example_azure_hypervisor.id
    hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example_azure_hypervisor_resource_pool.id
    azure_image_definition = {
        resource_group = "ExampleResourceGroup"
        use_image_gallery = true
        image_gallery_name = "ExampleImageGalleryName"
    }
}

resource "citrix_image_definition" "example_aws_ec2_image_definition" {
    name            = "Example AWS EC2 Image Definition"
    description     = "Example AWS EC2 Image Definition Description"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor      = citrix_aws_hypervisor.example_aws_hypervisor.id
    hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.example_aws_hypervisor_resource_pool.id
}

resource "citrix_image_definition" "example_vsphere_image_definition" {
    name            = "Example vSphere Image Definition"
    description     = "Example vSphere Image Definition Description"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor      = citrix_vsphere_hypervisor.example_vsphere_hypervisor.id
    hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.example_vsphere_hypervisor_resource_pool.id
}

resource "citrix_image_definition" "example_workspaces_core_image_definition" {
    name            = "Example Amazon WorkSpaces Core Image Definition"
    description     = "Example Amazon WorkSpaces Core Image Definition Description"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor      = citrix_amazon_workspaces_core_hypervisor.example-amazon-workspaces-core-hypervisor-using-api-key.id
    hypervisor_resource_pool = citrix_amazon_workspaces_core_hypervisor_resource_pool.example-amazon-workspaces-core-hypervisor-resource-pool.id
}
