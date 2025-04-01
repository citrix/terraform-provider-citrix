resource "citrix_image_definition" "example-image-definition" {
    name            = var.image_definition_name
    description     = "Description for example image definition"
    os_type         = "Windows"
    session_support = "MultiSession"
    hypervisor = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
    hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.example-vsphere-rp.id
}

resource "citrix_image_version" "example-image-version" {
    image_definition = citrix_image_definition.example-image-definition.id
    hypervisor = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
    hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.example-vsphere-rp.id
    description = "Description for example image version"

    vsphere_image_specs = {
        master_image_vm = var.vsphere_master_image_vm
        image_snapshot = var.vsphere_image_snapshot
        cpu_count = var.vsphere_cpu_count
        memory_mb = var.vsphere_memory_size
    }
}