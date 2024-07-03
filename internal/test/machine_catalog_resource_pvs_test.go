// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPvsMachineCatalogPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_PVS_MC_NAME"); v == "" {
		t.Fatal("TEST_PVS_MC_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_IDENTITY_TYPE"); v == "" {
		t.Fatal("TEST_PVS_IDENTITY_TYPE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_DOMAIN"); v == "" {
		t.Fatal("TEST_PVS_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_SERVICE_ACCOUNT"); v == "" {
		t.Fatal("TEST_PVS_SERVICE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_SERVICE_ACCOUNT_PASS"); v == "" {
		t.Fatal("TEST_PVS_SERVICE_ACCOUNT_PASS must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_MACHINE_PROFILE_VM_NAME"); v == "" {
		t.Fatal("TEST_PVS_MACHINE_PROFILE_VM_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_MACHINE_PROFILE_RG_NAME"); v == "" {
		t.Fatal("TEST_PVS_MACHINE_PROFILE_RG_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_NETWORK"); v == "" {
		t.Fatal("TEST_PVS_NETWORK must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_HYPERVISOR_ID"); v == "" {
		t.Fatal("TEST_PVS_HYPERVISOR_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_HYPERVISOR_RP_ID"); v == "" {
		t.Fatal("TEST_PVS_HYPERVISOR_RP_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_ZONE_ID"); v == "" {
		t.Fatal("TEST_PVS_ZONE_ID must be set for acceptance tests")
	}
}

func TestActiveDirectoryPvsMachineCatalogResourceAzure(t *testing.T) {
	name := os.Getenv("TEST_PVS_MC_NAME")
	isOnPremises := true
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildPvsCatalogResourceAzure(t, machinecatalog_testResources_azure_using_pvs),
					BuildPvsResource(t, pvs_test_data_source),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_PVS_SERVICE_ACCOUNT")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_PVS_NETWORK")),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testPvsMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildPvsCatalogResourceAzure(t, machinecatalog_testResources_azure_using_pvs_updated),
					BuildPvsResource(t, pvs_test_data_source),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "description", "updated description for pvs catalog"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// Delete machine test
			{
				Config: composeTestResourceTf(
					BuildPvsCatalogResourceAzure(t, machinecatalog_testResources_azure_using_pvs_delete_machine),
					BuildPvsResource(t, pvs_test_data_source),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "name", name),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testPvsMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

var machinecatalog_testResources_azure_using_pvs = `
resource "citrix_machine_catalog" "testPvsMachineCatalog" {
	name                		= "%s"
	description					= "description for pvs catalog"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "PVSStreaming"
	minimum_functional_level    = "L7_9"
	provisioning_scheme			= 	{
		hypervisor			 = "%s"
		hypervisor_resource_pool = "%s"
		identity_type = "%s"
		machine_domain_identity = {
			domain 						= "%s"
			service_account				= "%s"
			service_account_password 	= "%s"
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			azure_pvs_config = {
				pvs_site_id = "%s"
				pvs_vdisk_id = "%s"
			}
			machine_profile = {
				machine_profile_vm_name = "%s"
				machine_profile_resource_group = "%s"
			}
			storage_type = "Standard_LRS"
			use_managed_disks = true
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
			}
		}
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "test-machine-##"
			naming_scheme_type ="Numeric"
		}
	}

	zone						= "%s"
}`

var machinecatalog_testResources_azure_using_pvs_updated = `
resource "citrix_machine_catalog" "testPvsMachineCatalog" {
	name                		= "%s"
	description					= "updated description for pvs catalog"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "PVSStreaming"
	minimum_functional_level    = "L7_9"
	provisioning_scheme			= 	{
		hypervisor			 = "%s"
		hypervisor_resource_pool = "%s"
		identity_type = "%s"
		machine_domain_identity = {
			domain 						= "%s"
			service_account				= "%s"
			service_account_password 	= "%s"
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			azure_pvs_config = {
				pvs_site_id = "%s"
				pvs_vdisk_id = "%s"
			}
			machine_profile = {
				machine_profile_vm_name = "%s"
				machine_profile_resource_group = "%s"
			}
			storage_type = "Standard_LRS"
			use_managed_disks = true
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
			}
		}
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		number_of_total_machines = 	2
		machine_account_creation_rules ={
			naming_scheme =     "test-machine-##"
			naming_scheme_type ="Numeric"
		}
	}

	zone						= "%s"
}`

var machinecatalog_testResources_azure_using_pvs_delete_machine = `
resource "citrix_machine_catalog" "testPvsMachineCatalog" {
	name                		= "%s"
	description					= "updated description for pvs catalog"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "PVSStreaming"
	minimum_functional_level    = "L7_9"
	provisioning_scheme			= 	{
		hypervisor			 = "%s"
		hypervisor_resource_pool = "%s"
		identity_type = "%s"
		machine_domain_identity = {
			domain 						= "%s"
			service_account				= "%s"
			service_account_password 	= "%s"
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			azure_pvs_config = {
				pvs_site_id = "%s"
				pvs_vdisk_id = "%s"
			}
			machine_profile = {
				machine_profile_vm_name = "%s"
				machine_profile_resource_group = "%s"
			}
			storage_type = "Standard_LRS"
			use_managed_disks = true
			writeback_cache = {
				wbc_disk_storage_type = "Standard_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
			}
		}
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "test-machine-##"
			naming_scheme_type ="Numeric"
		}
	}

	zone						= "%s"
}`

func BuildPvsCatalogResourceAzure(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_PVS_MC_NAME")
	hypervisorId := os.Getenv("TEST_PVS_HYPERVISOR_ID")
	hypervisorResourcePoolId := os.Getenv("TEST_PVS_HYPERVISOR_RP_ID")
	identityType := os.Getenv("TEST_PVS_IDENTITY_TYPE")
	domainName := os.Getenv("TEST_PVS_DOMAIN")
	serviceAccount := os.Getenv("TEST_PVS_SERVICE_ACCOUNT")
	serviceAccountPass := os.Getenv("TEST_PVS_SERVICE_ACCOUNT_PASS")
	serviceOffering := os.Getenv("TEST_PVS_SERVICE_OFFERING")
	pvsSiteId := os.Getenv("TEST_PVS_SITE_ID")
	pvsVdiskId := os.Getenv("TEST_PVS_VDISK_ID")
	machineProfileVMName := os.Getenv("TEST_PVS_MACHINE_PROFILE_VM_NAME")
	machineProfileRGName := os.Getenv("TEST_PVS_MACHINE_PROFILE_RG_NAME")
	network := os.Getenv("TEST_PVS_NETWORK")
	zoneId := os.Getenv("TEST_PVS_ZONE_ID")

	return fmt.Sprintf(machineResource, name, hypervisorId, hypervisorResourcePoolId, identityType, domainName, serviceAccount, serviceAccountPass, serviceOffering, pvsSiteId, pvsVdiskId, machineProfileVMName, machineProfileRGName, network, zoneId)
}
