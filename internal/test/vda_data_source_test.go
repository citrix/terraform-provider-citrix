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

// TestVdaDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestVdaDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_VDA_DATASOURCE_MC_NAME"); v == "" {
		t.Fatal("TEST_VDA_DATASOURCE_MC_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_VDA_DATASOURCE_DG_NAME"); v == "" {
		t.Fatal("TEST_VDA_DATASOURCE_DG_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_VDA_DATASOURCE_MC_VDAS"); v == "" {
		t.Fatal("TEST_VDA_DATASOURCE_MC_VDAS must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_VDA_DATASOURCE_DG_VDAS"); v == "" {
		t.Fatal("TEST_VDA_DATASOURCE_DG_VDAS must be set for acceptance tests")
	}
}

func TestVdaDataSource(t *testing.T) {
	machineCatalog := os.Getenv("TEST_VDA_DATASOURCE_MC_NAME")
	deliveryGroup := os.Getenv("TEST_VDA_DATASOURCE_DG_NAME")
	machineCatalogVdas := os.Getenv("TEST_VDA_DATASOURCE_MC_VDAS")
	deliveryGroupVdas := os.Getenv("TEST_VDA_DATASOURCE_DG_VDAS")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestVdaDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Machine Catalog
			{
				Config: BuildVdaDataSource(t, vda_test_data_source_using_machine_catalog, machineCatalog),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the Name of the Machine Catalog
					resource.TestCheckResourceAttr("data.citrix_vda.test_vda_by_machine_catalog", "machine_catalog", machineCatalog),
					// Verify the list of VDAs in the Machine Catalog
					resource.TestCheckResourceAttr("data.citrix_vda.test_vda_by_machine_catalog", "vdas.#", strconv.Itoa(len(strings.Split(machineCatalogVdas, ",")))),
				),
			},
			// Read testing using Delivery Group
			{
				Config: BuildVdaDataSource(t, vda_test_data_source_using_delivery_group, deliveryGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the Name of the Delivery Group
					resource.TestCheckResourceAttr("data.citrix_vda.test_vda_by_delivery_group", "delivery_group", deliveryGroup),
					// Verify the list of VDAs in the Delivery Group
					resource.TestCheckResourceAttr("data.citrix_vda.test_vda_by_delivery_group", "vdas.#", strconv.Itoa(len(strings.Split(deliveryGroupVdas, ",")))),
				),
			},
		},
	})
}

func BuildVdaDataSource(t *testing.T, vdaDataSource string, nameOfDataSource string) string {
	return fmt.Sprintf(vdaDataSource, nameOfDataSource)
}

var (
	vda_test_data_source_using_machine_catalog = `
	data "citrix_vda" "test_vda_by_machine_catalog" {
		machine_catalog = "%s"
	}
	`

	vda_test_data_source_using_delivery_group = `
	data "citrix_vda" "test_vda_by_delivery_group" {
		delivery_group = "%s"
	}
	`
)
