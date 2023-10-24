package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMachineCatalogPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_MC_SERVICE_OFFERING"); v == "" {
		t.Fatal("TEST_MC_SERVICE_OFFERING must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_UPDATED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_RESOUCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_IMAGE_RESOUCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_IMAGE_STORAGE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_CONTAINER"); v == "" {
		t.Fatal("TEST_MC_IMAGE_CONTAINER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SUBNET"); v == "" {
		t.Fatal("TEST_MC_SUBNET must be set for acceptance tests")
	}
}

func TestMachineCatalogResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck(t)
			TestHypervisorResourcePoolPreCheck(t)
			TestMachineCatalogPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResource(t, machinecatalog_testResources),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", "test-catalog"),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT")),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.network_mapping.network", os.Getenv("TEST_MC_SUBNET")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"service_account", "service_account_password", "provisioning_scheme.network_mapping"},
			},
			//Update description, master image and add machine test
			{
				Config: BuildMachineCatalogResource(t, machinecatalog_testResources_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", "test-catalog"),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_config.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			// Delete machine test
			{
				Config: BuildMachineCatalogResource(t, machinecatalog_testResources_delete_machine),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", "test-catalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

var (
	machinecatalog_testResources = `
resource "citrix_daas_machine_catalog" "testMachineCatalog" {
	name                		= "test-catalog"
	description					= "on prem catalog for import testing"
	service_account				= "%s"
	service_account_password 	= "%s"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_scheme			= 	{
		machine_config = {
			hypervisor			 = citrix_daas_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool.id
			service_offering 	 = "%s"
			resource_group 		 = "%s"
            storage_account 	 = "%s"
            container 			 = "%s"
			master_image		 = "%s"
		}
		network_mapping = {
			network_device = "0"
			network 	   = "%s"
		}
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "test-machine-##"
			naming_scheme_type ="Numeric"
			domain =            "%s"
		}
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

	zone						= citrix_daas_zone.test.id
}
`
	machinecatalog_testResources_updated = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "test-catalog"
		description					= "updatedCatalog"
		service_account				= "%s"
		service_account_password 	= "%s"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_scheme			= 	{
			machine_config = {
				hypervisor			 = citrix_daas_hypervisor.testHypervisor.id
				hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool.id
				service_offering 	 = "%s"
				resource_group 		 = "%s"
				storage_account 	 = "%s"
				container 			 = "%s"
				master_image		 = "%s"
			}
			network_mapping = {
				network_device = "0"
				network 	   = "%s"
			}
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
				domain =            "%s"
			}
			storage_type = "Standard_LRS"
			use_managed_disks = true
			availability_zones = "1,3"
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
				storage_cost_saving = true
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`

	machinecatalog_testResources_delete_machine = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "test-catalog"
		description					= "updatedCatalog"
		service_account				= "%s"
		service_account_password 	= "%s"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_scheme			= 	{
			machine_config = {
				hypervisor			 = citrix_daas_hypervisor.testHypervisor.id
				hypervisor_resource_pool = citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool.id
				service_offering 	 = "%s"
				resource_group 		 = "%s"
				storage_account 	 = "%s"
				container 			 = "%s"
				master_image		 = "%s"
			}
			network_mapping = {
				network_device = "0"
				network 	   = "%s"
			}
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
				domain =            "%s"
			}
			storage_type = "Standard_LRS"
			use_managed_disks = true
			availability_zones = "1,3"
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
				storage_cost_saving = true
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`
)

func BuildMachineCatalogResource(t *testing.T, machineResource string) string {
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS")
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOUCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	subnet := os.Getenv("TEST_MC_SUBNET")
	if machineResource == machinecatalog_testResources_updated {
		master_image = os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")
	}

	//machine account
	domain := os.Getenv("TEST_MC_DOMAIN")

	return BuildHypervisorResourcePoolResource(t, hypervisor_resource_pool_testResource) + fmt.Sprintf(machineResource, service_account, service_account_pass, service_offering, resource_group, storage_account, container, master_image, subnet, domain)
}
