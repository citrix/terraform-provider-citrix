// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMachineCatalogDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestMachineCatalogDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_ID"); v == "" {
		t.Fatal("TEST_MACHINE_CATALOG_DATASOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_MACHINE_CATALOG_DATASOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_VDAS"); v == "" {
		t.Fatal("TEST_MACHINE_CATALOG_DATASOURCE_VDAS must be set for acceptance tests")
	}
}

func TestMachineCatalogDataSource(t *testing.T) {
	id := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_ID")
	name := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_NAME")
	vdas := os.Getenv("TEST_MACHINE_CATALOG_DATASOURCE_VDAS")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestMachineCatalogDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Machine Catalog
			{
				Config: BuildMachineCatalogDataSource(t, machine_catalog_test_data_source, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the Machine Catalog
					resource.TestCheckResourceAttr("data.citrix_machine_catalog.test_machine_catalog", "id", id),
					// Verify the list of VDAs in the Machine Catalog
					resource.TestCheckResourceAttr("data.citrix_machine_catalog.test_machine_catalog", "vdas.#", strconv.Itoa(len(strings.Split(vdas, ",")))),
				),
			},
		},
	})
}

func BuildMachineCatalogDataSource(t *testing.T, machineCatalogDataSource string, nameOfDataSource string) string {
	return fmt.Sprintf(machineCatalogDataSource, nameOfDataSource)
}

var (
	machine_catalog_test_data_source = `
	data "citrix_machine_catalog" "test_machine_catalog" {
		name = "%s"
	}
	`
)
