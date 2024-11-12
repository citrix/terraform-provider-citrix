// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMachinePropertiesResourcePreCheck validates the necessary env variable exist in the testing environment
func TestMachinePropertiesResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_NAME"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_NAME_UPDATED"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_NAME_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID_UPDATED"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_TAG"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_TAG must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_TAG_UPDATED"); v == "" {
		t.Fatal("TEST_MACHINE_PROPERTIES_RESOURCE_TAG_UPDATED must be set for acceptance tests")
	}
}

func TestMachinePropertiesResource(t *testing.T) {
	machineName := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_NAME")
	machineCatalogId := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID")
	tagId := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_TAG")

	machineName_updated := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_NAME_UPDATED")
	machineCatalogId_updated := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_MACHINE_CATALOG_ID_UPDATED")
	tagId_updated := os.Getenv("TEST_MACHINE_PROPERTIES_RESOURCE_TAG_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestMachinePropertiesResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildMachinePropertiesResourceWithoutTag(t, machineName, machineCatalogId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "name", machineName),
					// Verify the machine catalog id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "machine_catalog_id", machineCatalogId),
					// Verify the number of tags of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "tags.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_properties.test_machine_properties",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// add tag
			{
				Config: BuildMachinePropertiesResource(t, machineName, machineCatalogId, tagId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "name", machineName),
					// Verify the machine catalog id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "machine_catalog_id", machineCatalogId),
					// Verify the number of tags of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "tags.#", "1"),
					// Verify the tag id of the machine properties resource
					resource.TestCheckTypeSetElemAttr("citrix_machine_properties.test_machine_properties", "tags.*", tagId),
				),
			},
			// update machine id
			{
				Config: BuildMachinePropertiesResource(t, machineName_updated, machineCatalogId_updated, tagId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "name", machineName_updated),
					// Verify the machine catalog id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "machine_catalog_id", machineCatalogId_updated),
					// Verify the number of tags of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "tags.#", "1"),
					// Verify the tag id of the machine properties resource
					resource.TestCheckTypeSetElemAttr("citrix_machine_properties.test_machine_properties", "tags.*", tagId),
				),
			},
			// update tag
			{
				Config: BuildMachinePropertiesResource(t, machineName_updated, machineCatalogId_updated, tagId_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "name", machineName_updated),
					// Verify the machine catalog id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "machine_catalog_id", machineCatalogId_updated),
					// Verify the number of tags of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "tags.#", "1"),
					// Verify the tag id of the machine properties resource
					resource.TestCheckTypeSetElemAttr("citrix_machine_properties.test_machine_properties", "tags.*", tagId_updated),
				),
			},
			// remove tag
			{
				Config: BuildMachinePropertiesResourceWithoutTag(t, machineName_updated, machineCatalogId_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "name", machineName_updated),
					// Verify the machine catalog id of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "machine_catalog_id", machineCatalogId_updated),
					// Verify the number of tags of the machine properties resource
					resource.TestCheckResourceAttr("citrix_machine_properties.test_machine_properties", "tags.#", "0"),
				),
			},
		},
	})
}

func BuildMachinePropertiesResourceWithoutTag(t *testing.T, machineName string, machineCatalogId string) string {
	return fmt.Sprintf(machine_properties_test_resource_without_tag, machineName, machineCatalogId)
}

func BuildMachinePropertiesResource(t *testing.T, machineName string, machineCatalogId string, tagId string) string {
	return fmt.Sprintf(machine_properties_test_resource, machineName, machineCatalogId, tagId)
}

var (
	machine_properties_test_resource_without_tag = `
	resource "citrix_machine_properties" "test_machine_properties" {
		name = "%s"
		machine_catalog_id = "%s"
	}
	`

	machine_properties_test_resource = `
	resource "citrix_machine_properties" "test_machine_properties" {
		name = "%s"
		machine_catalog_id = "%s"
		tags = [ "%s" ]
	}
	`
)
