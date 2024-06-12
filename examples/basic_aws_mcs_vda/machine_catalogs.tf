resource "citrix_machine_catalog" "example-aws-catalog" {
    name                        = var.machine_catalog_name
    description                 = "Example multi-session catalog on AWS hypervisor"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type           = "MCS"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         =   {
        hypervisor               = citrix_aws_hypervisor.example-aws-hypervisor.id
        hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.example-aws-rp.id
        identity_type            = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = var.domain_fqdn
            domain_ou                = var.domain_ou
            service_account          = var.domain_service_account
            service_account_password = var.domain_service_account_password
        }
        aws_machine_config = {
            image_ami        = var.aws_ami_id
            master_image     = var.aws_ami_name
            service_offering = var.aws_service_offering
            security_groups  = [
                "default"
            ]
            tenancy_type     = "Shared"
        }
        number_of_total_machines =  1
        machine_account_creation_rules = {
            naming_scheme      = var.machine_catalog_naming_scheme
            naming_scheme_type = "Numeric"
        }
    }
}
