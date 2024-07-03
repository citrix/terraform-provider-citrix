// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestHypervisorDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestHypervisorDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_HYPERVISOR_DATASOURCE_ID"); v == "" {
		t.Fatal("TEST_HYPERVISOR_DATASOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_HYPERVISOR_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_HYPERVISOR_DATASOURCE_NAME must be set for acceptance tests")
	}
}

func TestHypervisorDataSource(t *testing.T) {
	id := os.Getenv("TEST_HYPERVISOR_DATASOURCE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Name
			{
				Config: BuildHypervisorDataSource(t, hypervisor_test_data_source_using_name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the hypervisor
					resource.TestCheckResourceAttr("data.citrix_hypervisor.test_hypervisor_by_name", "id", id),
				),
			},
		},
	})
}

func BuildHypervisorDataSource(t *testing.T, hypervisorDataSource string) string {
	hypervisorName := os.Getenv("TEST_HYPERVISOR_DATASOURCE_NAME")

	return fmt.Sprintf(hypervisorDataSource, hypervisorName)
}

var (
	hypervisor_test_data_source_using_name = `
	data "citrix_hypervisor" "test_hypervisor_by_name" {
		name = "%s"
	}
	`
)
