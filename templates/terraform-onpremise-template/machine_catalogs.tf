resource "citrix_daas_machine_catalog" "example-catalog" {
    name                        = "example-catalog"
    description                 = "description for example catalog"
    service_account             = "<Admin Username>"
    service_account_password    = "<Admin Password>"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = citrix_daas_zone.example-zone.id
    provisioning_scheme         = {
        storage_type = "Standard_LRS"
        use_managed_disks = true
        machine_config = {
            hypervisor = citrix_daas_hypervisor.example-azure-hypervisor.id
            hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.example-azure-rp.id
            service_offering = "Standard_D2_v2"
            resource_group = "<Resource Group for VDA image>"
            storage_account = "<Storage account for VDA image>"
            container = "<Container for VDA image>"
            master_image = "<Blob name for VDA image>"
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "ctx-pvdr-##"
            naming_scheme_type = "Numeric"
            domain = "<DomainFQDN>"
        }
    }
}