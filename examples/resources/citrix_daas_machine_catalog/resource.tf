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
            hypervisor = citrix_daas_hypervisor.azure-hypervisor-1.id
			hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.azure-hypervisor-resource-pool.id
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
			naming_scheme =     "multi-##"
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