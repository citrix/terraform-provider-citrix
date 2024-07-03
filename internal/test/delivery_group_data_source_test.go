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

// TestDeliveryGroupDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestDeliveryGroupDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_ID"); v == "" {
		t.Fatal("TEST_DELIVERY_GROUP_DATASOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_DELIVERY_GROUP_DATASOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_VDAS"); v == "" {
		t.Fatal("TEST_DELIVERY_GROUP_DATASOURCE_VDAS must be set for acceptance tests")
	}
}

func TestDeliveryGroupDataSource(t *testing.T) {
	id := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_ID")
	name := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_NAME")
	vdas := os.Getenv("TEST_DELIVERY_GROUP_DATASOURCE_VDAS")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestDeliveryGroupDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Delivery Group
			{
				Config: BuildDeliveryGroupDataSource(t, delivery_group_test_data_source, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the Delivery Group
					resource.TestCheckResourceAttr("data.citrix_delivery_group.test_delivery_group", "id", id),
					// Verify the list of VDAs in the Delivery Group
					resource.TestCheckResourceAttr("data.citrix_delivery_group.test_delivery_group", "vdas.#", strconv.Itoa(len(strings.Split(vdas, ",")))),
				),
			},
		},
	})
}

func BuildDeliveryGroupDataSource(t *testing.T, deliveryGroupDataSource string, nameOfDataSource string) string {
	return fmt.Sprintf(deliveryGroupDataSource, nameOfDataSource)
}

var (
	delivery_group_test_data_source = `
	data "citrix_delivery_group" "test_delivery_group" {
		name = "%s"
	}
	`
)
