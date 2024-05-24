resource "citrix_machine_catalog" "example-azure-mtsession" {
	name                		= "example-azure-mtsession"
	description					= "Example multi-session catalog on Azure hypervisor"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "MCS"
	minimum_functional_level    = "L7_20"
	provisioning_scheme			= 	{
		hypervisor = citrix_azure_hypervisor.example-azure-hypervisor.id
		hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example-azure-hypervisor-resource-pool.id
		identity_type      = "ActiveDirectory"
		machine_domain_identity = {
            domain                   = "<DomainFQDN>"
			domain_ou				 = "<DomainOU>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
		azure_machine_config = {
			storage_type = "Standard_LRS"
			use_managed_disks = true
            service_offering = "Standard_D2_v2"
            azure_machine_config = {
                resource_group = "<Azure resource group name for image vhd>"
                storage_account = "<Azure storage account name for image vhd>"
                container = "<Azure storage container for image vhd>"
                master_image = "<Image vhd blob name>"
            }
			writeback_cache = {
				wbc_disk_storage_type = "pd-standard"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
                writeback_cache_memory_size_mb = 256
				storage_cost_saving = true
			}
        }
		network_mapping = [
            {
                network_device = "0"
                network = "<Azure Subnet for machine>"
            }
        ]
		availability_zones = "1,2,..."
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "az-multi-##"
			naming_scheme_type ="Numeric"
		}
	}
}

resource "citrix_machine_catalog" "example-aws-mtsession" {
    name                        = "example-aws-mtsession"
    description                 = "Example multi-session catalog on AWS hypervisor"
   	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "MCS"
    provisioning_scheme         = {
		hypervisor = citrix_aws_hypervisor.example-aws-hypervisor.id
		hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.example-aws-hypervisor-resource-pool.id
		identity_type      = "ActiveDirectory"
		machine_domain_identity = {
            domain                   = "<DomainFQDN>"
			domain_ou				 = "<DomainOU>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        aws_machine_config = {
            image_ami = "<AMI ID for VDA>"
			master_image = "<Image template AMI name>"
			service_offering = "t2.small"
            security_groups = [
                "default"
            ]
            tenancy_type = "Shared"
        }
		network_mapping = [
            {
                network_device = "0"
                network = "10.0.128.0/20"
            }
        ]
		number_of_total_machines =  1
        machine_account_creation_rules ={
			naming_scheme 	   = "aws-multi-##"
			naming_scheme_type = "Numeric"
        }
    }	
}

resource "citrix_machine_catalog" "example-gcp-mtsession" {
    name                        = "example-gcp-mtsession"
    description                 = "Example multi-session catalog on GCP hypervisor"
   	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "MCS"
    provisioning_scheme         = {
		hypervisor = citrix_gcp_hypervisor.example-gcp-hypervisor.id
		hypervisor_resource_pool = citrix_gcp_hypervisor_resource_pool.example-gcp-hypervisor-resource-pool.id
		identity_type      = "ActiveDirectory"
		machine_domain_identity = {
            domain                   = "<DomainFQDN>"
			domain_ou				 = "<DomainOU>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        gcp_machine_config = {
            
            machine_profile = "<Machine profile template VM name>"
            master_image = "<Image template VM name>"
            machine_snapshot = "<Image template VM snapshot name>"
			storage_type = "pd-standard"
			writeback_cache = {
				wbc_disk_storage_type = "pd-standard"
				persist_wbc = true
				persist_os_disk = true
				writeback_cache_disk_size_gb = 127
                writeback_cache_memory_size_mb = 256
			}
        }
		availability_zones = "{project name}:{region}:{availability zone1},{project name}:{region}:{availability zone2},..."
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "gcp-multi-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-vsphere-mtsession" {
    name                        = "example-vsphere-mtsession"
    description                 = "Example multi-session catalog on Vsphere hypervisor"
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
        hypervisor = citrix_vsphere_hypervisor.vsphere-hypervisor-1.id
        hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.vsphere-hypervisor-rp-1.id
        vsphere_machine_config = {
            master_image_vm = "<Image VM name>"
            image_snapshot = "<Snapshot 1>/<Snapshot 2>/<Snapshot 3>/..."
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

resource "citrix_machine_catalog" "example-xenserver-mtsession" {
    name                        = "example-xenserver-mtsession"
    description                 = "Example multi-session catalog on XenServer hypervisor"
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
        hypervisor = citrix_xenserver_hypervisor.xenserver-hypervisor-1.id
        hypervisor_resource_pool = citrix_xenserver_hypervisor_resource_pool.xenserver-hypervisor-rp-1.id
        xenserver_machine_config = {
            master_image_vm = "<Image VM name>"
            image_snapshot = "<Snapshot 1>/<Snapshot 2>/<Snapshot 3>/..."
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

resource "citrix_machine_catalog" "example-nutanix-mtsession" {
    name                        = "example-nutanix-mtsession"
    description                 = "Example multi-session catalog on Nutanix hypervisor"
    provisioning_type 			= "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = citrix_zone.example-zone.id
    provisioning_scheme         = {
        identity_type = "ActiveDirectory"
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
        hypervisor = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id
        hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.example-nutanix-rp.id
        nutanix_machine_config = {
            container = "<Container name>"
            master_image = "<Image name>"
            cpu_count = 2
            memory_mb = 4096
            cores_per_cpu_count = 2
        }
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
    }
}

resource "citrix_machine_catalog" "example-manual-power-managed-mtsession" {
	name                		= "example-manual-power-managed-mtsession"
	description					= "Example manual power managed multi-session catalog"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	is_power_managed			= true
	is_remote_pc 			  	= false
	provisioning_type 			= "Manual"
	machine_accounts = [
        {
            hypervisor = citrix_azure_hypervisor.example-azure-hypervisor.id
            machines = [
                {
                    region = "East US"
                    resource_group_name = "machine-resource-group-name"
                    machine_name = "Domain\\MachineName"
                }
            ]
        }
    ]
}

resource "citrix_machine_catalog" "example-manual-non-power-managed-mtsession" {
	name                		= "example-manual-non-power-managed-mtsession"
	description					= "Example manual non power managed multi-session catalog"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	is_power_managed			= false
	is_remote_pc 			  	= false
	provisioning_type 			= "Manual"
	machine_accounts = [
        {
            machines = [
                {
                    machine_name = "Domain\\MachineName1"
                },
				{
                    machine_name = "Domain\\MachineName2"
                }
            ]
        }
    ]
}

resource "citrix_machine_catalog" "example-remote-pc" {
	name                		= "example-remote-pc-catalog"
	description					= "Example Remote PC catalog"
	zone						= "<zone Id>"
	allocation_type				= "Static"
	session_support				= "SingleSession"
	is_power_managed			= false
	is_remote_pc 			  	= true
	provisioning_type 			= "Manual"
	machine_accounts = [
        {
            machines = [
                {
                    machine_name = "Domain\\MachineName1"
                },
				{
                    machine_name = "Domain\\MachineName2"
                }
            ]
        }
    ]
	remote_pc_ous = [
        {
            include_subfolders = false
            ou_name = "OU=Example OU,DC=domain,DC=com"
        }
    ]
}

resource "citrix_machine_catalog" "example-non-domain-joined-azure-mcs" {
	name                		= "example-non-domain-joined-azure-mcs"
	description					= "Example catalog on Azure without domain join"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type 			= "MCS"
	provisioning_scheme			= 	{
		hypervisor = citrix_azure_hypervisor.example-azure-hypervisor.id
		hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.example-azure-hypervisor-resource-pool.id
		identity_type      = "Workgroup" # Workgroup specifies that the machines are not domain-joined
		# Example using Azure, other hypervisors can be used as well
        azure_machine_config = {
			storage_type = "Standard_LRS"
			use_managed_disks = true
            service_offering = "Standard_D2_v2"
            azure_master_image = {
                resource_group 		 = "<Azure resource group name for image vhd>"
				storage_account 	 = "<Azure storage account name for image vhd>"
				container 			 = "<Azure storage container for image vhd>"
                master_image = "<Image vhd blob name>"
            }
			writeback_cache = {
				wbc_disk_storage_type = "pd-standard"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
                writeback_cache_memory_size_mb = 256
				storage_cost_saving = true
			}
        }
		availability_zones = "1,2,..."
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "ndj-multi-##"
			naming_scheme_type ="Numeric"
		}
	}
}