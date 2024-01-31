resource "citrix_daas_machine_catalog" "example-catalog" {
    name                        = "example-catalog"
    description                 = "description for example catalog"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "MCS"
    zone                        = citrix_daas_zone.example-zone.id
    provisioning_scheme         = {
        hypervisor               = citrix_daas_azure_hypervisor.example-azure-hypervisor.id
        hypervisor_resource_pool = citrix_daas_azure_hypervisor_resource_pool.example-azure-rp.id
        identity_type            = "ActiveDirectory"
        machine_domain_identity = {
                domain                   = "<DomainFQDN>"
                service_account          = "<Admin Username>"
                service_account_password = "<Admin Password>"
            }
        azure_machine_config     = {
            service_offering    = "Standard_D2_v2"
            resource_group      = "<Resource Group for VDA image>"
            storage_account     = "<Storage account for VDA image>"
            container           = "<Container for VDA image>"
            master_image        = "<Blob name for VDA image>"
            storage_type        = "Standard_LRS"
            use_managed_disks   = true
        }
        number_of_total_machines = 1
        network_mapping = {
            network_device = "0"
            network = "<Azure Subnet for machine>"
        }
        machine_account_creation_rules = {
            naming_scheme = "ctx-pvdr-##"
            naming_scheme_type = "Numeric"
        }
    }
}