// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMachineCatalogPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME"); v == "" {
		t.Fatal("TEST_MC_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS must be set for acceptance tests")
	}
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

func TestMachineCatalogResourceAzure(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT")),
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
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password"},
			},
			//Update description, master image and add machine test
			{
				Config: BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.azure_machine_config.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			// Delete machine test
			{
				Config: BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_GCP(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_GCP"); v == "" {
		t.Fatal("TEST_MC_NAME_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_GCP"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_STORAGE_TYPE_GCP"); v == "" {
		t.Fatal("TEST_MC_STORAGE_TYPE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_AVAILABILITY_ZONES_GCP"); v == "" {
		t.Fatal("TEST_MC_AVAILABILITY_ZONES_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_PROFILE_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_PROFILE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_GCP"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_SNAPSHOT_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_SNAPSHOT_GCP must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MC_DOMAIN_GCP"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_Subnet_GCP"); v == "" {
		t.Fatal("TEST_MC_Subnet_GCP must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceGCP(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_GCP")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_GCP(t)
			TestHypervisorResourcePoolPreCheck_GCP(t)
			TestMachineCatalogPreCheck_GCP(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceGCP(t, machinecatalog_testResources_gcp),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.availability_zones", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password"},
			},
			//Update description, master image and add machine test
			{
				Config: BuildMachineCatalogResourceGCP(t, machinecatalog_testResources_gcp_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Mannual_Power_Managed_Azure(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_REGION_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_Azure"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_Azure must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Azure(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Mannual_Power_Managed_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceManualPowerManagedAzure(t, machinecatalog_testResources_manual_power_managed_azure),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_GCP(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP"); v == "" {
		t.Fatal("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_GCP must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_GCP(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_GCP(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_GCP(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceManualPowerManagedGCP(t, machinecatalog_testResources_manual_power_managed_gcp),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_daas_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Mannual_Non_Power_Managed(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Non_Power_Managed(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Mannual_Non_Power_Managed(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceManualNonPowerManaged(t, machinecatalog_testResources_manual_non_power_managed),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogNonManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogNonManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalogNonManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_machine_catalog.testMachineCatalogNonManualPowerManaged",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_RemotePC(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_NAME_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_OU_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_OU_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_RemotePC(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_REMOTE_PC")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestMachineCatalogPreCheck_RemotePC(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildMachineCatalogResourceRemotePC(t, machinecatalog_testResources_remote_pc),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "name", name),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_daas_machine_catalog.testMachineCatalog", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_daas_machine_catalog.testMachineCatalog",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

var (
	machinecatalog_testResources_azure = `
resource "citrix_daas_machine_catalog" "testMachineCatalog" {
	name                		= "%s"
	description					= "on prem catalog for import testing"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "MCS"
	provisioning_scheme			= 	{
		hypervisor			 = citrix_daas_azure_hypervisor.testHypervisor.id
		hypervisor_resource_pool = citrix_daas_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
		identity_type = "ActiveDirectory"
		machine_domain_identity = {
			domain 						= "%s"
			service_account				= "%s"
			service_account_password 	= "%s"
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			resource_group 		 = "%s"
            storage_account 	 = "%s"
            container 			 = "%s"
			master_image		 = "%s"
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
		network_mapping = {
			network_device = "0"
			network 	   = "%s"
		}
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "test-machine-##"
			naming_scheme_type ="Numeric"
		}
	}

	zone						= citrix_daas_zone.test.id
}
`
	machinecatalog_testResources_azure_updated = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_daas_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_daas_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 = "%s"
				resource_group 		 = "%s"
				storage_account 	 = "%s"
				container 			 = "%s"
				master_image		 = "%s"
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
			network_mapping = {
				network_device = "0"
				network 	   = "%s"
			}
			availability_zones = "1,3"
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`

	machinecatalog_testResources_azure_delete_machine = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "updatedCatalog"		
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 	 = citrix_daas_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_daas_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 	 = "%s"
				resource_group 		 	 = "%s"
				storage_account 	 	 = "%s"
				container 			 	 = "%s"
				master_image		 	 = "%s"
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
			network_mapping = {
				network_device = "0"
				network 	   = "%s"
			}
			availability_zones = "1,3"
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`

	machinecatalog_testResources_gcp = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "on prem catalog for import testing"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_daas_gcp_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_daas_gcp_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			gcp_machine_config = {
				storage_type = "%s"
				machine_profile = "%s"
				master_image		 = "%s"
				machine_snapshot = "%s"
			}
			number_of_total_machines = 	1
			availability_zones = "%s"
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`

	machinecatalog_testResources_gcp_updated = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_daas_gcp_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_daas_gcp_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			gcp_machine_config = {
				storage_type = "%s"
				machine_profile = "%s"
				master_image		 = "%s"
				machine_snapshot = "%s"
			}
			number_of_total_machines = 	2
			availability_zones = "%s"
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_daas_zone.test.id
	}
	`

	machinecatalog_testResources_manual_power_managed_azure = `
	resource "citrix_daas_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                		= "%s"
		description					= "manual power managed multi-session catalog testing for Azure Hypervisor"
		zone						= citrix_daas_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= true
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				hypervisor = citrix_daas_azure_hypervisor.testHypervisor.id
				machines = [
					{
						region = "%s"
						resource_group_name = "%s"
						machine_name = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_gcp = `
	resource "citrix_daas_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                		= "%s"
		description					= "manual power managed multi-session catalog testing"
		zone						= citrix_daas_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= true
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				hypervisor = citrix_daas_gcp_hypervisor.testHypervisor.id
				machines = [
					{
						region = "%s"
						project_name = "%s"
						machine_name = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_non_power_managed = `
	resource "citrix_daas_machine_catalog" "testMachineCatalogNonManualPowerManaged" {
		name                		= "%s"
		description					= "manual non power managed multi-session catalog testing"
		zone						= citrix_daas_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= false
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				machines = [
					{
						machine_name = "%s"
					}
				]
			}
		]
	}
	`
	machinecatalog_testResources_remote_pc = `
	resource "citrix_daas_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "on prem catalog for import testing remotePC"
		allocation_type				= "%s"
		session_support				= "SingleSession"
		provisioning_type			= "Manual"
		is_remote_pc				= true
		is_power_managed			= false
		machine_accounts			= [
			{
				machines = [
					{
						machine_name = "%s"
					}
				]
			}
		]
		remote_pc_ous = [
			{
				include_subfolders = %s
				ou_name = "%s"
			}
		]
		zone						= citrix_daas_zone.test.id
	}
	`
)

func BuildMachineCatalogResourceAzure(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS")
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOUCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	subnet := os.Getenv("TEST_MC_SUBNET")
	if machineResource == machinecatalog_testResources_azure_updated {
		master_image = os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")
	}

	//machine account
	domain := os.Getenv("TEST_MC_DOMAIN")

	return BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure) + fmt.Sprintf(machineResource, name, domain, service_account, service_account_pass, service_offering, resource_group, storage_account, container, master_image, subnet)
}

func BuildMachineCatalogResourceGCP(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_GCP")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_GCP")
	storage_type := os.Getenv("TEST_MC_STORAGE_TYPE_GCP")
	availability_zones := os.Getenv("TEST_MC_AVAILABILITY_ZONES_GCP")
	machine_profile := os.Getenv("TEST_MC_MACHINE_PROFILE_GCP")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_GCP")
	machine_snapshot := os.Getenv("TEST_MC_MACHINE_SNAPSHOT_GCP")

	//machine account
	domain := os.Getenv("TEST_MC_DOMAIN_GCP")

	return BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp) + fmt.Sprintf(machineResource, name, domain, service_account, service_account_pass, storage_type, machine_profile, master_image, machine_snapshot, availability_zones)
}

func BuildMachineCatalogResourceManualPowerManagedAzure(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_Azure")
	region := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED")
	resource_group := os.Getenv("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return BuildHypervisorResourceAzure(t, hypervisor_testResources) + fmt.Sprintf(machineResource, name, allocation_type, session_support, region, resource_group, machine_name)
}

func BuildMachineCatalogResourceManualPowerManagedGCP(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_GCP")
	region := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP")
	project_name := os.Getenv("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp) + fmt.Sprintf(machineResource, name, allocation_type, session_support, region, project_name, machine_name)
}

func BuildMachineCatalogResourceManualNonPowerManaged(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_NON_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED")

	return BuildZoneResource(t, zone_testResource) + fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_name)
}

func BuildMachineCatalogResourceRemotePC(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_REMOTE_PC")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_REMOTE_PC")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_REMOTE_PC")
	ou := os.Getenv("TEST_MC_OU_REMOTE_PC")
	include_subfolders := os.Getenv("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC")

	return BuildZoneResource(t, zone_testResource) + fmt.Sprintf(machineResource, name, allocation_type, machine_name, include_subfolders, ou)
}
