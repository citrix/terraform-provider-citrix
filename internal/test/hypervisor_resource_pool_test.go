// Copyright © 2023. Citrix Systems, Inc.

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

func TestHypervisorResourcePoolPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_REGION"); v == "" {
		t.Fatal("TEST_HYPERV_RP_REGION must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_HYPERV_RP_VIRTUAL_NETWORK_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK"); v == "" {
		t.Fatal("TEST_HYPERV_RP_VIRTUAL_NETWORK must be set for acceptance tests")
	}
	if v := os.Getenv("Test_HYPERV_RP_SUBNETS"); v == "" {
		t.Fatal("Test_HYPERV_RP_SUBNETS must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolAzureRM(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of virtual network resource group name
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "virtual_network_resource_group", os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK_RESOURCE_GROUP")),
					// Verify name of virtual network
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "virtual_network", os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK")),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "region", os.Getenv("TEST_HYPERV_RP_REGION")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "subnets.#", strconv.Itoa(len(strings.Split(os.Getenv("Test_HYPERV_RP_SUBNETS"), ",")))),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"subnet"},
			},
			// Update and Read
			{
				Config: BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_GCP(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_GCP"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_REGION_GCP"); v == "" {
		t.Fatal("TEST_HYPERV_RP_REGION_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("Test_HYPERV_RP_SUBNETS_GCP"); v == "" {
		t.Fatal("Test_HYPERV_RP_SUBNETS_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_PROJECT_NAME_GCP"); v == "" {
		t.Fatal("TEST_HYPERV_RP_PROJECT_NAME_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_VPC_GCP"); v == "" {
		t.Fatal("TEST_HYPERV_RP_VPC_GCP must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolGCP(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_GCP")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_GCP(t)
			TestHypervisorResourcePoolPreCheck_GCP(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "region", os.Getenv("TEST_HYPERV_RP_REGION_GCP")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "subnets.#", strconv.Itoa(len(strings.Split(os.Getenv("Test_HYPERV_RP_SUBNETS_GCP"), ",")))),
					// Verify name of the project
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "project_name", os.Getenv("TEST_HYPERV_RP_PROJECT_NAME_GCP")),
					// Verify name of the vpc
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "vpc", os.Getenv("TEST_HYPERV_RP_VPC_GCP")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId_GCP,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_updated_testResource_gcp),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_Xenserver(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NETWORK_1_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("Test_HYPERV_RP_NETWORK_2_XENSERVER"); v == "" {
		t.Fatal("Test_HYPERV_RP_NETWORK_2_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_RP_STORAGE_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolXenserver(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Xenserver(t)
			TestHypervisorResourcePoolPreCheck_Xenserver(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourcePoolResourceXenServer(t),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "storage.0", os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")),
					// Verify name of the project
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.0", os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId_XenServer,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: BuildHypervisorResourcePoolResourceXenServerUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "2"),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_Vsphere(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_DATACENTER_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_DATACENTER_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_HOST_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_HOST_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NETWORK_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_STORAGE_1_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_STORAGE_2_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_STORAGE_2_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolVsphere(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Vsphere(t)
			TestHypervisorResourcePoolPreCheck_Vsphere(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourcePoolResourceVsphere(t),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "2"),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE")),
					// Verify name of the project
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE")),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:             true,
				ImportStateIdFunc:       generateImportStateId_Vsphere,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster"},
			},
			// Update and Read
			{
				Config: BuildHypervisorResourcePoolResourceVsphereUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "1"),
				),
			},
		},
	})
}

func generateImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool"
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

func generateImportStateId_GCP(state *terraform.State) (string, error) {
	resourceName := "citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool"
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

func generateImportStateId_XenServer(state *terraform.State) (string, error) {
	resourceName := "citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool"
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

func generateImportStateId_Vsphere(state *terraform.State) (string, error) {
	resourceName := "citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool"
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
	hypervisor_resource_pool_testResource_azure = `
resource "citrix_azure_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "%s"
	hypervisor = citrix_azure_hypervisor.testHypervisor.id
    region = "%s"
	virtual_network_resource_group = "%s"
	virtual_network = "%s"
	subnets = %s
}
`

	hypervisor_resource_pool_updated_testResource_azure = `
resource "citrix_azure_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "%s-updated"
	hypervisor = citrix_azure_hypervisor.testHypervisor.id
    region = "%s"
	virtual_network_resource_group = "%s"
	virtual_network = "%s"
	subnets = %s
}
`
	hypervisor_resource_pool_testResource_gcp = `
resource "citrix_gcp_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s"
	hypervisor = citrix_gcp_hypervisor.testHypervisor.id
	project_name = "%s"
	region = "%s"
	subnets = %s
	vpc = "%s"
}
`
	hypervisor_resource_pool_updated_testResource_gcp = `
resource "citrix_gcp_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_gcp_hypervisor.testHypervisor.id
	project_name = "%s"
	region = "%s"
	subnets = %s
	vpc = "%s"
}	
`

	hypervisor_resource_pool_testResource_xenserver = `
resource "citrix_xenserver_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s"
	hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
	networks = ["%s"]
	storage = ["%s"]
	temporary_storage = ["%s"]
}
`
	hypervisor_resource_pool_updated_testResource_xenserver = `
resource "citrix_xenserver_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
	networks = ["%s", "%s"]
	storage = ["%s"]
	temporary_storage = ["%s"]
}	
`

	hypervisor_resource_pool_testResource_vsphere = `
resource "citrix_vsphere_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s"
	hypervisor = citrix_vsphere_hypervisor.testHypervisor.id
	cluster = {
		datacenter = "%s"
		host = "%s"
	}
	networks = ["%s"]
	storage = [{
		storage_name = "%s"
	},
	{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
}
`
	hypervisor_resource_pool_updated_testResource_vsphere = `
resource "citrix_vsphere_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_vsphere_hypervisor.testHypervisor.id
	cluster = {
		datacenter = "%s"
		host = "%s"
	}
	networks = ["%s"]
	storage = [{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
}	
`
)

func BuildHypervisorResourcePoolResourceAzure(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME")
	region := os.Getenv("TEST_HYPERV_RP_REGION")
	virtualNetworkResourceGroup := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK_RESOURCE_GROUP")
	virtualNetwork := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK")
	subnet := os.Getenv("Test_HYPERV_RP_SUBNETS")

	return BuildHypervisorResourceAzure(t, hypervisor_testResources) + fmt.Sprintf(hypervisorRP, name, region, virtualNetworkResourceGroup, virtualNetwork, subnet)
}

func BuildHypervisorResourcePoolResourceGCP(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_GCP")
	region := os.Getenv("TEST_HYPERV_RP_REGION_GCP")
	subnet := os.Getenv("Test_HYPERV_RP_SUBNETS_GCP")
	projectName := os.Getenv("TEST_HYPERV_RP_PROJECT_NAME_GCP")
	vpc := os.Getenv("TEST_HYPERV_RP_VPC_GCP")

	return BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp) + fmt.Sprintf(hypervisorRP, name, projectName, region, subnet, vpc)
}

func BuildHypervisorResourcePoolResourceXenServer(t *testing.T) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER")
	network1 := os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")
	storage := os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")

	return BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver) + fmt.Sprintf(hypervisor_resource_pool_testResource_xenserver, name, network1, storage, tempStorage)
}

func BuildHypervisorResourcePoolResourceXenServerUpdated(t *testing.T) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER")
	network1 := os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")
	network2 := os.Getenv("TEST_HYPERV_RP_NETWORK_2_XENSERVER")
	storage := os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")

	return BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver) + fmt.Sprintf(hypervisor_resource_pool_updated_testResource_xenserver, name, network1, network2, storage, tempStorage)
}

func BuildHypervisorResourcePoolResourceVsphere(t *testing.T) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE")
	datacenter := os.Getenv("TEST_HYPERV_RP_DATACENTER_VSPHERE")
	host := os.Getenv("TEST_HYPERV_RP_HOST_VSPHERE")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE")
	storage_2 := os.Getenv("TEST_HYPERV_RP_STORAGE_2_VSPHERE")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE")

	return BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere) + fmt.Sprintf(hypervisor_resource_pool_testResource_vsphere, name, datacenter, host, network, storage_1, storage_2, tempStorage)
}

func BuildHypervisorResourcePoolResourceVsphereUpdated(t *testing.T) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE")
	datacenter := os.Getenv("TEST_HYPERV_RP_DATACENTER_VSPHERE")
	host := os.Getenv("TEST_HYPERV_RP_HOST_VSPHERE")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE")

	return BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere) + fmt.Sprintf(hypervisor_resource_pool_updated_testResource_vsphere, name, datacenter, host, network, storage_1, tempStorage)
}
