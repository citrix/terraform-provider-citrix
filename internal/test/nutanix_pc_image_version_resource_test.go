// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var nutanixPCImageVersionTestVariables = []string{
	"TEST_NUTANIX_PC_IMAGE_DEFINITION_NAME",
	"TEST_NUTANIX_PC_IMAGE_VERSION_CLUSTER",
	"TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE",
	"TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE_VERSION",
	"TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE",
	"TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE_VERSION",
	"TEST_NUTANIX_PC_IMAGE_VERSION_NETWORK",
}

func TestNutanixPCImageVersionPreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, nutanixPCImageVersionTestVariables)
}

func TestNutanixPCImageVersionResource(t *testing.T) {
	zoneInput := os.Getenv("TEST_ZONE_INPUT_NUTANIX_PC")
	resourceName := "citrix_image_version.test_nutanix_pc_image_version"
	var imageVersionId, imageVersionNumber string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_NutanixPC(t)
			TestHypervisorResourcePoolPreCheck_NutanixPC(t)
			TestNutanixPCImageVersionPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildNutanixPCImageVersion(t, 8192),
					BuildHypervisorResourcePoolResourceNutanixPC(t, hypervisor_resource_pool_testResource_nutanix_pc),
					BuildHypervisorResourceNutanixPC(t, hypervisor_testResources_nutanix_pc),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", captureResourceAttr(&imageVersionId)),
					resource.TestCheckResourceAttrWith(resourceName, "version_number", captureResourceAttr(&imageVersionNumber)),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.cluster", os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_CLUSTER")),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.master_image", os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE")),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.master_image_version", os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE_VERSION")),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.cores_per_cpu_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.memory_mb", "8192"),
					resource.TestCheckResourceAttr(resourceName, "network_mapping.#", "1"),
				),
			},
			{
				// network_mapping is refreshed for Prism Central even on import, so nothing is ignored.
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: generateNutanixPCImageVersionImportStateId,
			},
			{
				Config: composeTestResourceTf(
					BuildNutanixPCImageVersion(t, 16384),
					BuildHypervisorResourcePoolResourceNutanixPC(t, hypervisor_resource_pool_testResource_nutanix_pc),
					BuildHypervisorResourceNutanixPC(t, hypervisor_testResources_nutanix_pc),
					BuildZoneResource(t, zoneInput, false),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &imageVersionId),
					resource.TestCheckResourceAttrPtr(resourceName, "version_number", &imageVersionNumber),
					resource.TestCheckResourceAttr(resourceName, "nutanix_pc_image_specs.memory_mb", "16384"),
				),
			},
		},
	})
}

func generateNutanixPCImageVersionImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_image_version.test_nutanix_pc_image_version"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["image_definition"], rawState["id"]), nil
}

func BuildNutanixPCImageVersion(t *testing.T, memoryMb int) string {
	return fmt.Sprintf(nutanixPCImageVersionTestResource,
		os.Getenv("TEST_NUTANIX_PC_IMAGE_DEFINITION_NAME"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_NETWORK"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_CLUSTER"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MASTER_IMAGE_VERSION"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE_VERSION"),
		memoryMb,
	)
}

var nutanixPCImageVersionTestResource = `
resource "citrix_image_definition" "test_nutanix_pc_image_definition" {
	name                     = "%s"
	description              = "Nutanix Prism Central image definition for acceptance testing"
	os_type                  = "Windows"
	session_support          = "MultiSession"
	hypervisor               = citrix_nutanix_pc_hypervisor.testHypervisor.id
	hypervisor_resource_pool = citrix_nutanix_pc_hypervisor_resource_pool.testHypervisorResourcePool.id
}

resource "citrix_image_version" "test_nutanix_pc_image_version" {
	image_definition         = citrix_image_definition.test_nutanix_pc_image_definition.id
	hypervisor               = citrix_nutanix_pc_hypervisor.testHypervisor.id
	hypervisor_resource_pool = citrix_nutanix_pc_hypervisor_resource_pool.testHypervisorResourcePool.id
	description              = "Nutanix Prism Central image version for acceptance testing"
	network_mapping = [
		{
			network_device = "0"
			network        = "%s"
		}
	]
	nutanix_pc_image_specs = {
		cluster                 = "%s"
		master_image            = "%s"
		master_image_version    = "%s"
		machine_profile         = "%s"
		machine_profile_version = "%s"
		cpu_count               = 2
		cores_per_cpu_count     = 2
		memory_mb               = %d
	}
}
`

func TestMachineCatalogResourceNutanixPCPreparedImage(t *testing.T) {
	zoneInput := os.Getenv("TEST_ZONE_INPUT_NUTANIX_PC")
	resourceName := "citrix_machine_catalog.testMachineCatalogPreparedImage"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_NutanixPC(t)
			TestHypervisorResourcePoolPreCheck_NutanixPC(t)
			TestNutanixPCImageVersionPreCheck(t)
			TestMachineCatalogPreCheck_NutanixPC(t)
		},
		Steps: []resource.TestStep{
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceNutanixPCPreparedImage(t, 1),
					BuildNutanixPCImageVersion(t, 8192),
					BuildHypervisorResourcePoolResourceNutanixPC(t, hypervisor_resource_pool_testResource_nutanix_pc),
					BuildHypervisorResourceNutanixPC(t, hypervisor_testResources_nutanix_pc),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_scheme.nutanix_pc_machine_config.prepared_image.image_version", "citrix_image_version.test_nutanix_pc_image_version", "id"),
					resource.TestCheckNoResourceAttr(resourceName, "provisioning_scheme.nutanix_pc_machine_config.master_image"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceNutanixPCPreparedImage(t, 2),
					BuildNutanixPCImageVersion(t, 8192),
					BuildHypervisorResourcePoolResourceNutanixPC(t, hypervisor_resource_pool_testResource_nutanix_pc),
					BuildHypervisorResourceNutanixPC(t, hypervisor_testResources_nutanix_pc),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
		},
	})
}

func BuildMachineCatalogResourceNutanixPCPreparedImage(t *testing.T, machineCount int) string {
	return fmt.Sprintf(machineCatalogNutanixPCPreparedImageTestResource,
		os.Getenv("TEST_MC_NAME_NUTANIX_PC")+"-prep",
		machineCount,
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_CLUSTER"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_MACHINE_PROFILE_VERSION"),
		os.Getenv("TEST_MC_MEMORY_MB_NUTANIX_PC"),
		os.Getenv("TEST_MC_CPU_COUNT_NUTANIX_PC"),
		os.Getenv("TEST_MC_CORES_PER_CPU_COUNT_NUTANIX_PC"),
		os.Getenv("TEST_NUTANIX_PC_IMAGE_VERSION_NETWORK"),
		os.Getenv("TEST_MC_SERVICE_ACCOUNT_NUTANIX_PC"),
		os.Getenv("TEST_MC_DOMAIN_NUTANIX_PC"),
		os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_NUTANIX_PC"),
	)
}

var machineCatalogNutanixPCPreparedImageTestResource = `
resource "citrix_machine_catalog" "testMachineCatalogPreparedImage" {
	name                = "%s"
	description         = "nutanix prism central prepared image catalog for acceptance testing"
	provisioning_type   = "MCS"
	allocation_type     = "Random"
	session_support     = "MultiSession"
	zone                = citrix_zone.test.id
	provisioning_scheme = {
		identity_type            = "ActiveDirectory"
		number_of_total_machines = %d
		machine_account_creation_rules = {
			naming_scheme      = "test-pcp-##"
			naming_scheme_type = "Numeric"
		}
		hypervisor               = citrix_nutanix_pc_hypervisor.testHypervisor.id
		hypervisor_resource_pool = citrix_nutanix_pc_hypervisor_resource_pool.testHypervisorResourcePool.id
		nutanix_pc_machine_config = {
			cluster = "%s"
			prepared_image = {
				image_definition = citrix_image_definition.test_nutanix_pc_image_definition.id
				image_version    = citrix_image_version.test_nutanix_pc_image_version.id
			}
			machine_profile         = "%s"
			machine_profile_version = "%s"
			memory_mb               = "%s"
			cpu_count               = "%s"
			cores_per_cpu_count     = "%s"
		}
		network_mapping = [
			{
				network_device = "0"
				network        = "%s"
			}
		]
		machine_domain_identity = {
			service_account          = "%s"
			domain                   = "%s"
			service_account_password = "%s"
		}
	}
}
`
