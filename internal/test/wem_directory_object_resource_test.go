// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDirectoryObject(t *testing.T) {
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestWemSiteResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: composeTestResourceTf(
					BuildDirectoryObjectResource(t, wem_directory_object_test_resource),
					BuildWemSiteResource(t),
					BuildMachineCatalogResourceWorkgroup(t, machinecatalog_testResources_workgroup),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify if the directory object is created and enabled
					resource.TestCheckResourceAttr("citrix_wem_directory_object.test_wem_directory", "enabled", "true"),
				),
			},
			// Import test
			{
				ResourceName:            "citrix_wem_directory_object.test_wem_directory",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update and Read test
			{
				Config: composeTestResourceTf(
					BuildDirectoryObjectResource(t, wem_directory_object_test_resource_updated),
					BuildWemSiteResource(t),
					BuildMachineCatalogResourceWorkgroup(t, machinecatalog_testResources_workgroup),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify if the directory object is disabled
					resource.TestCheckResourceAttr("citrix_wem_directory_object.test_wem_directory", "enabled", "false"),
				),
			},
		},
	})
}

func BuildDirectoryObjectResource(t *testing.T, directoryResource string) string {
	return directoryResource
}

var (
	wem_directory_object_test_resource = `
	resource "citrix_wem_directory_object" "test_wem_directory" {
		configuration_set_id = citrix_wem_configuration_set.test_wem_site.id
		machine_catalog_id = citrix_machine_catalog.testMachineCatalog-WG.id
		enabled = true
	}
	`
)

var (
	wem_directory_object_test_resource_updated = `
	resource "citrix_wem_directory_object" "test_wem_directory" {
		configuration_set_id = citrix_wem_configuration_set.test_wem_site.id
		machine_catalog_id = citrix_machine_catalog.testMachineCatalog-WG.id
		enabled = false
	}
	`
)
