// Copyright © 2024. Citrix Systems, Inc.

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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),

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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp),
					BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_GCP")),
				),

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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_updated_testResource_gcp),
					BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_GCP")),
				),
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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceXenServer(t, hypervisor_resource_pool_testResource_xenserver),
					BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_XENSERVER")),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")),
					// Verify name of the project
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")),
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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceXenServerUpdated(t, hypervisor_resource_pool_updated_testResource_xenserver),
					BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_XENSERVER")),
				),
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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceVsphere(t, hypervisor_resource_pool_testResource_vsphere),
					BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_VSPHERE")),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "1"),
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
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceVsphereUpdated(t, hypervisor_resource_pool_updated_testResource_vsphere),
					BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_VSPHERE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "2"),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_Nutanix(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_NETWORK_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NETWORK_NUTANIX must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolNutanix(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_NUTANIX")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Nutanix(t)
			TestHypervisorResourcePoolPreCheck_Nutanix(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceNutanix(t, hypervisor_resource_pool_testResource_nutanix),
					BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_NUTANIX")),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_NUTANIX")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId_Nutanix,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceNutanix(t, hypervisor_resource_pool_updated_testResource_nutanix),
					BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_NUTANIX")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_SCVMM(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_HOST_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_HOST_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_NETWORK_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NETWORK_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_STORAGE_1_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_STORAGE_1_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_STORAGE_2_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_STORAGE_2_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_RP_TEMP_STORAGE_SCVMM must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolSCVMM(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_SCVMM")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_SCVMM(t)
			TestHypervisorResourcePoolPreCheck_SCVMM(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceSCVMM(t, hypervisor_resource_pool_testResource_scvmm),
					BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_SCVMM")),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "networks.#", "1"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "networks.0", os.Getenv("TEST_HYPERV_RP_NETWORK_SCVMM")),
					// Verify subnets
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_STORAGE_1_SCVMM")),
					// Verify name of the project
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.#", "1"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "temporary_storage.0.storage_name", os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_SCVMM")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId_SCVMM,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceSCVMMUpdated(t, hypervisor_resource_pool_updated_testResource_scvmm),
					BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_SCVMM")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool", "storage.#", "2"),
				),
			},
		},
	})
}

func TestHypervisorResourcePoolPreCheck_Aws_Ec2(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_RP_NAME_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_RP_NAME_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_VPC_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_RP_VPC_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_RP_AVAILABILITY_ZONE_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_RP_AVAILABILITY_ZONE_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("Test_HYPERV_RP_SUBNETS_AWS_EC2"); v == "" {
		t.Fatal("Test_HYPERV_RP_SUBNETS_AWS_EC2 must be set for acceptance tests")
	}
}

func TestHypervisorResourcePoolAwsEc2(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_RP_NAME_AWS_EC2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_AWS_EC2(t)
			TestHypervisorResourcePoolPreCheck_Aws_Ec2(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAwsEc2(t, hypervisor_resource_pool_testResource_aws_ec2),
					BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AWS_EC2")),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool", "name", name),
					// Verify name of the region
					resource.TestCheckResourceAttr("citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool", "vpc", os.Getenv("TEST_HYPERV_RP_VPC_AWS_EC2")),
					resource.TestCheckResourceAttr("citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool", "availability_zone", os.Getenv("TEST_HYPERV_RP_AVAILABILITY_ZONE_AWS_EC2")),
					resource.TestCheckResourceAttr("citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool", "subnets.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool",
				ImportState:       true,
				ImportStateIdFunc: generateImportStateId_Aws_Ec2,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAwsEc2(t, hypervisor_resource_pool_updated_testResource_aws_ec2),
					BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AWS_EC2")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", name)),
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

func generateImportStateId_Nutanix(state *terraform.State) (string, error) {
	resourceName := "citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool"
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

func generateImportStateId_SCVMM(state *terraform.State) (string, error) {
	resourceName := "citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool"
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

func generateImportStateId_Aws_Ec2(state *terraform.State) (string, error) {
	resourceName := "citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool"
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
	storage = [
	{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
}
`
	hypervisor_resource_pool_updated_testResource_xenserver = `
resource "citrix_xenserver_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
	networks = ["%s", "%s"]
	storage = [
	{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
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
	storage = [
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
	},
	{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
}	
`
	hypervisor_resource_pool_testResource_scvmm = `
resource "citrix_scvmm_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s"
	hypervisor = citrix_scvmm_hypervisor.testHypervisor.id
	host = "%s"
	networks = ["%s"]
	storage = [
	{
		storage_name = "%s"
	}]
	temporary_storage = [{
		storage_name = "%s"
	}]
}
`
	hypervisor_resource_pool_updated_testResource_scvmm = `
resource "citrix_scvmm_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_scvmm_hypervisor.testHypervisor.id
	host = "%s"
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

	hypervisor_resource_pool_testResource_nutanix = `
resource "citrix_nutanix_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s"
	hypervisor = citrix_nutanix_hypervisor.testHypervisor.id
	networks = ["%s"]
}
`

	hypervisor_resource_pool_updated_testResource_nutanix = `
resource "citrix_nutanix_hypervisor_resource_pool" "testHypervisorResourcePool" {
	name = "%s-updated"
	hypervisor = citrix_nutanix_hypervisor.testHypervisor.id
	networks = ["%s"]
}
`

	hypervisor_resource_pool_testResource_aws_ec2 = `
resource "citrix_aws_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "%s"
	hypervisor = citrix_aws_hypervisor.testHypervisor.id
    vpc = "%s"
	availability_zone = "%s"
	subnets = %s
}
`

	hypervisor_resource_pool_updated_testResource_aws_ec2 = `
resource "citrix_aws_hypervisor_resource_pool" "testHypervisorResourcePool" {
    name = "%s-updated"
	hypervisor = citrix_aws_hypervisor.testHypervisor.id
    vpc = "%s"
	availability_zone = "%s"
	subnets = %s
}
`
)

func BuildHypervisorResourcePoolResourceAzure(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME")
	region := os.Getenv("TEST_HYPERV_RP_REGION")
	virtualNetworkResourceGroup := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK_RESOURCE_GROUP")
	virtualNetwork := os.Getenv("TEST_HYPERV_RP_VIRTUAL_NETWORK")
	subnet := os.Getenv("Test_HYPERV_RP_SUBNETS")

	return fmt.Sprintf(hypervisorRP, name, region, virtualNetworkResourceGroup, virtualNetwork, subnet)
}

func BuildHypervisorResourcePoolResourceGCP(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_GCP")
	region := os.Getenv("TEST_HYPERV_RP_REGION_GCP")
	subnet := os.Getenv("Test_HYPERV_RP_SUBNETS_GCP")
	projectName := os.Getenv("TEST_HYPERV_RP_PROJECT_NAME_GCP")
	vpc := os.Getenv("TEST_HYPERV_RP_VPC_GCP")

	return fmt.Sprintf(hypervisorRP, name, projectName, region, subnet, vpc)
}

func BuildHypervisorResourcePoolResourceXenServer(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER")
	network1 := os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")
	storage := os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")

	return fmt.Sprintf(hypervisorRP, name, network1, storage, tempStorage)
}

func BuildHypervisorResourcePoolResourceXenServerUpdated(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_XENSERVER")
	network1 := os.Getenv("TEST_HYPERV_RP_NETWORK_1_XENSERVER")
	network2 := os.Getenv("TEST_HYPERV_RP_NETWORK_2_XENSERVER")
	storage := os.Getenv("TEST_HYPERV_RP_STORAGE_XENSERVER")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_XENSERVER")

	return fmt.Sprintf(hypervisorRP, name, network1, network2, storage, tempStorage)
}

func BuildHypervisorResourcePoolResourceVsphere(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE")
	datacenter := os.Getenv("TEST_HYPERV_RP_DATACENTER_VSPHERE")
	host := os.Getenv("TEST_HYPERV_RP_HOST_VSPHERE")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE")

	return fmt.Sprintf(hypervisorRP, name, datacenter, host, network, storage_1, tempStorage)
}

func BuildHypervisorResourcePoolResourceVsphereUpdated(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_VSPHERE")
	datacenter := os.Getenv("TEST_HYPERV_RP_DATACENTER_VSPHERE")
	host := os.Getenv("TEST_HYPERV_RP_HOST_VSPHERE")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_VSPHERE")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_VSPHERE")
	storage_2 := os.Getenv("TEST_HYPERV_RP_STORAGE_2_VSPHERE")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_VSPHERE")

	return fmt.Sprintf(hypervisorRP, name, datacenter, host, network, storage_1, storage_2, tempStorage)
}

func BuildHypervisorResourcePoolResourceNutanix(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_NUTANIX")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_NUTANIX")

	return fmt.Sprintf(hypervisorRP, name, network)
}

func BuildHypervisorResourcePoolResourceSCVMM(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_SCVMM")
	host := os.Getenv("TEST_HYPERV_RP_HOST_SCVMM")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_SCVMM")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_SCVMM")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_SCVMM")

	return fmt.Sprintf(hypervisorRP, name, host, network, storage_1, tempStorage)
}

func BuildHypervisorResourcePoolResourceSCVMMUpdated(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_SCVMM")
	host := os.Getenv("TEST_HYPERV_RP_HOST_SCVMM")
	network := os.Getenv("TEST_HYPERV_RP_NETWORK_SCVMM")
	storage_1 := os.Getenv("TEST_HYPERV_RP_STORAGE_1_SCVMM")
	storage_2 := os.Getenv("TEST_HYPERV_RP_STORAGE_2_SCVMM")
	tempStorage := os.Getenv("TEST_HYPERV_RP_TEMP_STORAGE_SCVMM")

	return fmt.Sprintf(hypervisorRP, name, host, network, storage_1, storage_2, tempStorage)
}

func BuildHypervisorResourcePoolResourceAwsEc2(t *testing.T, hypervisorRP string) string {
	name := os.Getenv("TEST_HYPERV_RP_NAME_AWS_EC2")
	vpc := os.Getenv("TEST_HYPERV_RP_VPC_AWS_EC2")
	availability_zone := os.Getenv("TEST_HYPERV_RP_AVAILABILITY_ZONE_AWS_EC2")
	subnets := os.Getenv("Test_HYPERV_RP_SUBNETS_AWS_EC2")

	return fmt.Sprintf(hypervisorRP, name, vpc, availability_zone, subnets)
}
