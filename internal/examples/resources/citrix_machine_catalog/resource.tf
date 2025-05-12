resource "citrix_machine_catalog" "example-azure-mtsession" {
	name                		= "example-azure-mtsession"
	description					= "Example multi-session catalog on Azure hypervisor"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type 			= "MCS"
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
            azure_master_image = {
                # shared_subscription = var.azure_image_subscription # Uncomment if the image is from a subscription outside of the hypervisor's subscription

                # Resource Group is required for any type of Azure master image
                resource_group       = var.azure_resource_group

                # For Azure master image from managed disk or snapshot
                master_image         = var.azure_master_image

                # For Azure image gallery
                # gallery_image = {
                #     gallery    = var.azure_gallery_name
                #     definition = var.azure_gallery_image_definition
                #     version    = var.azure_gallery_image_version
                # }
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
		availability_zones = ["1","2"]
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "az-multi-##"
			naming_scheme_type ="Numeric"
		}
	}
}

resource "citrix_machine_catalog" "example_azure_prepared_image_mtsession" {
	name                		= "example_azure_prepared_image_mtsession"
	description					= "Example multi-session catalog on Azure hypervisor with Prepared Image"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type 			= "MCS"
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
            prepared_image = {
                image_definition = citrix_image_definition.example_image_definition.id
                image_version    = citrix_image_version.example_azure_image_version.id
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
		availability_zones = ["1","2"]
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
    description                 = "Example multi-session catalog on vSphere hypervisor"
    zone                        = "<zone Id>"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type 			= "MCS"
    provisioning_scheme         = {
        hypervisor = citrix_vsphere_hypervisor.vsphere-hypervisor-1.id
        hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.vsphere-hypervisor-rp-1.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        vsphere_machine_config = {
            master_image_vm = "<Image VM name>"
            image_snapshot = "<Snapshot 1>/<Snapshot 2>/<Snapshot 3>/..."
            cpu_count = 2
            memory_mb = 4096
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-vsphere-prepared-image" {
    name                        = "example-vsphere-prepared-image"
    description                 = "Example multi-session catalog on vSphere hypervisor with prepared image"
    zone                        = "<zone Id>"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type 			= "MCS"
    provisioning_scheme         = {
        hypervisor = citrix_vsphere_hypervisor.vsphere-hypervisor-1.id
        hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.vsphere-hypervisor-rp-1.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        vsphere_machine_config = {
            prepared_image = {
                image_definition = citrix_image_definition.example_vsphere_image_definition.id
                image_version    = citrix_image_version.example_vsphere_image_version.id
            }
            cpu_count = 2
            memory_mb = 4096
            machine_profile = "<machine profile template name>"
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-xenserver-mtsession" {
    name                        = "example-xenserver-mtsession"
    description                 = "Example multi-session catalog on XenServer hypervisor"
    zone                        = "<zone Id>"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type 			= "MCS"
    provisioning_scheme         = {
        hypervisor = citrix_xenserver_hypervisor.xenserver-hypervisor-1.id
        hypervisor_resource_pool = citrix_xenserver_hypervisor_resource_pool.xenserver-hypervisor-rp-1.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        xenserver_machine_config = {
            master_image_vm = "<Image VM name>"
            image_snapshot = "<Snapshot 1>/<Snapshot 2>/<Snapshot 3>/..."
            cpu_count = 2
            memory_mb = 4096
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-nutanix-mtsession" {
    name                        = "example-nutanix-mtsession"
    description                 = "Example multi-session catalog on Nutanix hypervisor"
    zone                        = citrix_zone.example-zone.id
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type 			= "MCS"
    provisioning_scheme         = {
        hypervisor = citrix_nutanix_hypervisor.example-nutanix-hypervisor.id
        hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.example-nutanix-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        nutanix_machine_config = {
            container = "<Container name>"
            master_image = "<Image name>"
            cpu_count = 2
            memory_mb = 4096
            cores_per_cpu_count = 2
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-scvmm-mtsession" {
    name                        = "example-scvmm-mtsession"
    description                 = "Example multi-session catalog on SCVMM hypervisor"
    provisioning_type = "MCS"
    allocation_type             = "Random"
    session_support             = "MultiSession"
    zone                        = citrix_zone.scvmm-zone.id
    provisioning_scheme         = {
        hypervisor = citrix_scvmm_hypervisor.example-scvmm-hypervisor.id
        hypervisor_resource_pool = citrix_scvmm_hypervisor_resource_pool.example-scvmm-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        scvmm_machine_config = {
            master_image = "<master image>"
            cpu_count = 1
            memory_mb = 2048
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
    }
}

resource "citrix_machine_catalog" "example-openshift-mtsession" {
    name                        = "example-openshift-mtsession"
    description                 = "Example multi-session catalog on OpenShift hypervisor"
    zone                        = citrix_zone.openshift-zone.id
    allocation_type             = "Random"
    session_support             = "MultiSession"
    provisioning_type             = "MCS"
    provisioning_scheme         = {
        hypervisor = citrix_openshift_hypervisor.example-openshift-hypervisor.id
        hypervisor_resource_pool = citrix_openshift_hypervisor_resource_pool.example-openshift-rp.id
        identity_type = "ActiveDirectory"
        machine_domain_identity = {
            domain                   = "<DomainFQDN>"
            service_account          = "<Admin Username>"
            service_account_password = "<Admin Password>"
        }
        openshift_machine_config = {
            master_image_vm = "<Image VM name>"
            cpu_count = 2
            memory_mb = 4096
            writeback_cache = {
                writeback_cache_disk_size_gb = 32
                writeback_cache_memory_size_mb = 2048
            }
        }
        
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "catalog-##"
            naming_scheme_type = "Numeric"
        }
        network_mapping = [{
            network = "<network name>"
            network_device = "0"
        }]
    }
}

resource "citrix_machine_catalog" "example-azure-pvs-mtsession" {
	name                		= "example-azure-pvs-mtsession"
	description					= "Example multi-session catalog on Azure hypervisor"
	zone						= "<zone Id>"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type 			= "PVSStreaming"
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
            azure_pvs_config = {
                pvs_site_id = data.citrix_pvs.example_pvs_config.pvs_site_id
				pvs_vdisk_id = data.citrix_pvs.example_pvs_config.pvs_vdisk_id
            }
			use_managed_disks = true
            service_offering = "Standard_D2_v2"
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
                writeback_cache_memory_size_mb = 256
			}
        }
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "az-pvs-multi-##"
			naming_scheme_type ="Numeric"
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
                    machine_account = "domain\\machine-name"
                    machine_name = "MachineName"
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
                    machine_account = "DOMAIN\\MachineName1"
                },
				{
                    machine_account = "DOMAIN\\MachineName2"
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
                    machine_account = "DOMAIN\\MachineName1"
                },
				{
                    machine_account = "DOMAIN\\MachineName2"
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
                # shared_subscription = var.azure_image_subscription # Uncomment if the image is from a subscription outside of the hypervisor's subscription

                # Resource Group is required for any type of Azure master image
                resource_group       = var.azure_resource_group

                # For Azure master image from managed disk or snapshot
                master_image         = var.azure_master_image

                # For Azure image gallery
                # gallery_image = {
                #     gallery    = var.azure_gallery_name
                #     definition = var.azure_gallery_image_definition
                #     version    = var.azure_gallery_image_version
                # }
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
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "ndj-multi-##"
			naming_scheme_type ="Numeric"
		}
	}
}