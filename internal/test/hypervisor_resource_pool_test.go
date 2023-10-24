package test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestHypervisorResourcePoolPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_REGION"); v == "" {
		t.Fatal("TEST_HYPERV_REGION must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_HYPERV_VIRTUAL_NETWORK_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK"); v == "" {
		t.Fatal("TEST_HYPERV_VIRTUAL_NETWORK must be set for acceptance tests")
	}
	if v := os.Getenv("Test_HYPERV_SUBNETS"); v == "" {
		t.Fatal("Test_HYPERV_SUBNETS must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolAzureRM(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck(t)
			TestHypervisorResourcePoolPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourcePoolResource(t, hypervisor_resource_pool_testResource),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "name", "test-hypervisor-resource-pool"),
					// Verify name of virtual network resource group name
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "virtual_network_resource_group", os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK_RESOURCE_GROUP")),
					// Verify name of virtual network
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "virtual_network", os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK")),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "region", os.Getenv("TEST_HYPERV_REGION")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "subnets.#", strconv.Itoa(len(strings.Split(os.Getenv("Test_HYPERV_SUBNETS"), ",")))),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"subnet"},
			},
			// Update and Read
			{
				Config: BuildHypervisorResourcePoolResource(t, hypervisor_resource_pool_updated_testResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool", "name", "test-hypervisor-resource-pool-updated"),
				),
			},
		},
	})
}

func generateImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_daas_hypervisor_resource_pool.testHypervisorResourcePool"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["hypervisor"], rawState["id"]), nil
}

var (
	hypervisor_resource_pool_testResource = `
resource "citrix_daas_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "test-hypervisor-resource-pool"
	hypervisor = citrix_daas_hypervisor.testHypervisor.id
    region = "%s"
	virtual_network_resource_group = "%s"
	virtual_network = "%s"
	subnets = %s
}
`
)

var (
	hypervisor_resource_pool_updated_testResource = `
resource "citrix_daas_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "test-hypervisor-resource-pool-updated"
	hypervisor = citrix_daas_hypervisor.testHypervisor.id
    region = "%s"
	virtual_network_resource_group = "%s"
	virtual_network = "%s"
	subnets = %s
}
`
)

func BuildHypervisorResourcePoolResource(t *testing.T, hypervisor string) string {
	region := os.Getenv("TEST_HYPERV_REGION")
	virtualNetworkResourceGroup := os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK_RESOURCE_GROUP")
	virtualNetwork := os.Getenv("TEST_HYPERV_VIRTUAL_NETWORK")
	subnet := os.Getenv("Test_HYPERV_SUBNETS")

	return BuildHypervisorResource(t) + fmt.Sprintf(hypervisor, region, virtualNetworkResourceGroup, virtualNetwork, subnet)
}
