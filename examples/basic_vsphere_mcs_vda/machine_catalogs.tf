resource "citrix_machine_catalog" "example-catalog" {
    name                        = "example-catalog"
    description                 = "description for example catalog"
    provisioning_type 			= "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = "<zone Id>"
    provisioning_scheme         = {
        identity_type = "ActiveDirectory"
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
        hypervisor = citrix_vsphere_hypervisor.example-vsphere-hypervisor.id
        hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.example-vsphere-rp.id
        vsphere_machine_config = {
            master_image_vm = "<Image VM or snapshot name>"
            cpu_count = 2
            memory_mb = 4096
        }
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
    }
}