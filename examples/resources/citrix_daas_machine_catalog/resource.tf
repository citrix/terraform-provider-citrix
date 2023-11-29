resource "citrix_daas_machine_catalog" "example-azure-mtsession" {
	name                		= "example-azure-mtsession"
	description					= "Example multi-session catalog on Azure hypervisor"
	zone						= "{zone Id}"
	service_account				= "{domain-admin-account}"
	service_account_password 	= "{domain-admin-password}"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_scheme			= 	{
		machine_config = {
            hypervisor = citrix_daas_hypervisor.example-azure-hypervisor.id
			hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.example-azure-hypervisor-resource-pool.id
            service_offering = "Standard_D2_v2"
            resource_group = "{Azure resource group name for image vhd}"
            storage_account = "{Azure storage account name for image vhd}"
            container = "{Azure storage container for image vhd}"
            master_image = "{Image vhd blob name}"
        }
		network_mapping = {
            network_device = "0"
            network = "{Azure Subnet for machine}"
        }
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "az-multi-##"
			naming_scheme_type ="Numeric"
			domain =            "{domain-fqdn}"
		}
		os_type = "Windows"
		storage_type = "Standard_LRS"
		use_managed_disks = true
		writeback_cache = {
			wbc_disk_storage_type = "Standard_LRS"
			persist_wbc = true
			persist_os_disk = true
			persist_vm = true
			writeback_cache_disk_size_gb = 127
			storage_cost_saving = true
		}
	}
}

resource "citrix_daas_machine_catalog" "example-gcp-mtsession" {
    name                        = "example-gcp-mtsession"
    description                 = "Example multi-session catalog on GCP hypervisor"
   	zone						= "{zone Id}"
	service_account				= "{domain-admin-account}"
	service_account_password 	= "{domain-admin-password}"
	allocation_type				= "Random"
	session_support				= "MultiSession"
    provisioning_scheme         = {
        storage_type = "pd-standard"
        availability_zones = "{project name}:{region}:{availability zone1},{project name}:{region}:{availability zone2},..."
        machine_config = {
            hypervisor = citrix_daas_hypervisor.example-gcp-hypervisor.id
            hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.example-gcp-hypervisor-resource-pool.id
            machine_profile = "{Machine profile template VM name}"
            master_image = "{Image template VM name}"
            machine_snapshot = "{Image template VM snapshot name}"
        }
        number_of_total_machines = 1
        machine_account_creation_rules = {
            naming_scheme = "gcp-multi-##"
            naming_scheme_type = "Numeric"
            domain = "serenity.local"
        }
		writeback_cache = {
			wbc_disk_storage_type = "Standard_LRS"
			persist_wbc = true
			persist_os_disk = true
			persist_vm = true
			writeback_cache_disk_size_gb = 127
			storage_cost_saving = true
		}
    }
}