resource "citrix_machine_catalog" "example-aws-catalog" {
    name                        = "example-aws-catalog"
    description                 = "Example multi-session catalog on AWS hypervisor"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    is_power_managed            = true
    is_remote_pc                = false
    provisioning_type           = "MCS"
    zone                        = citrix_zone.example-zone.id
    minimum_functional_level    = "L7_20"
    provisioning_scheme         =   {
        hypervisor               = citrix_aws_hypervisor.example-aws-hypervisor.id
        hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.example-aws-rp.id
        identity_type            = "ActiveDirectory"
        machine_domain_identity  = {
            domain                   = "<DomainFQDN>"
            domain_ou                = "<DomainOU>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        aws_machine_config = {
            image_ami        = "<AWS AMI ID>"
            master_image     = "<AWS AMI Name>"
            service_offering = "t2.small"
        }
        number_of_total_machines =  1
        network_mapping = {
            network_device = "0"
            network        = "<AWS Subnet Mask>"
        }
        machine_account_creation_rules = {
            naming_scheme      = "ctx-aws-###"
            naming_scheme_type = "Numeric"
        }
    }
}
