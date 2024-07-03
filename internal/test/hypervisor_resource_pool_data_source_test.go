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

// TestHypervisorResourcePoolDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestHypervisorResourcePoolDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_ID"); v == "" {
		t.Fatal("TEST_HYPERVISOR_RP_DATASOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_NAME"); v == "" {
		t.Fatal("TEST_HYPERVISOR_RP_DATASOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_NETWORKS"); v == "" {
		t.Fatal("TEST_HYPERVISOR_RP_DATASOURCE_NETWORKS must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolDataSource(t *testing.T) {
	id := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_ID")
	networks := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_NETWORKS")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorDataSourcePreCheck(t)
			TestHypervisorResourcePoolDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Name
			{
				Config: BuildHypervisorResourcePoolDataSource(t, hypervisor_resource_pool_test_data_source_using_name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the resource pool
					resource.TestCheckResourceAttr("data.citrix_hypervisor_resource_pool.test_resource_pool_by_name", "id", id),
					// Verify the networks of the resource pool
					resource.TestCheckResourceAttr("data.citrix_hypervisor_resource_pool.test_resource_pool_by_name", "networks.#", strconv.Itoa(len(strings.Split(networks, ",")))),
				),
			},
		},
	})
}

func BuildHypervisorResourcePoolDataSource(t *testing.T, resourcePoolDataSource string) string {
	resourcePoolName := os.Getenv("TEST_HYPERVISOR_RP_DATASOURCE_NAME")
	hypervisorName := os.Getenv("TEST_HYPERVISOR_DATASOURCE_NAME")

	return fmt.Sprintf(resourcePoolDataSource, resourcePoolName, hypervisorName)
}

var (
	hypervisor_resource_pool_test_data_source_using_name = `
	data "citrix_hypervisor_resource_pool" "test_resource_pool_by_name" {
		name = "%s"
		hypervisor_name = "%s"
	}
	`
)
