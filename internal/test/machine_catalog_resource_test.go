// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_machine_catalog", &resource.Sweeper{
		Name: "citrix_machine_catalog",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			isOnPremises := (client.ClientConfig.CustomerId == "CitrixOnPremises")

			var errs *multierror.Error
			// MCS AD machine catalog sweep
			machineCatalogName := os.Getenv("TEST_MC_NAME")
			err := machineCatalogSweeper(ctx, machineCatalogName, client)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			// MCS Hybrid Azure AD machine catalog sweep
			err = machineCatalogSweeper(ctx, machineCatalogName+"-HybAAD", client)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			if !isOnPremises {
				// MCS Azure AD machine catalog sweep
				err = machineCatalogSweeper(ctx, machineCatalogName+"-AAD", client)
				if err != nil {
					errs = multierror.Append(errs, err)
				}

				// MCS Workgroup machine catalog sweep
				err = machineCatalogSweeper(ctx, machineCatalogName+"-WRKGRP", client)
				if err != nil {
					errs = multierror.Append(errs, err)
				}
			}

			// Manual machine catalog sweep
			machineCatalogName = os.Getenv("TEST_MC_NAME_MANUAL")
			err = machineCatalogSweeper(ctx, machineCatalogName, client)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			// Remote PC machine catalog sweep
			machineCatalogName = os.Getenv("TEST_MC_NAME_REMOTE_PC")
			err = machineCatalogSweeper(ctx, machineCatalogName, client)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			return errs.ErrorOrNil()
		},
		Dependencies: []string{"citrix_delivery_group"},
	})
}

func TestMachineCatalogPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME"); v == "" {
		t.Fatal("TEST_MC_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN"); v == "" {
		t.Fatal("TEST_MC_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_OFFERING"); v == "" {
		t.Fatal("TEST_MC_SERVICE_OFFERING must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_UPDATED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_IMAGE_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_IMAGE_STORAGE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_CONTAINER"); v == "" {
		t.Fatal("TEST_MC_IMAGE_CONTAINER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SUBNET"); v == "" {
		t.Fatal("TEST_MC_SUBNET must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_VDA_AZURE_RESOURCE_GROUP must be set for acceptance tests")
	}
}
func TestActiveDirectoryMachineCatalogResourceAzure(t *testing.T) {
	ctx := context.Background()
	client := sharedClientForSweepers(ctx)

	orchestrationApiVersion := client.ClientConfig.OrchestrationApiVersion

	name := os.Getenv("TEST_MC_NAME")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestServiceAccountPreCheck_AD(t)
			TestMachineCatalogPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: func() string {
					var mcResource string
					if orchestrationApiVersion >= 120 {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "-AD", "ActiveDirectory")
					} else {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_without_serv_acc_uid, "-AD", "ActiveDirectory")
					}
					return composeTestResourceTf(
						mcResource,
						BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated),
						BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
						BuildHypervisorResourceAzure(t, hypervisor_testResources),
						BuildZoneResource(t, zoneInput, false),
					)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-AD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "-AD", "ActiveDirectory"),
					BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.azure_machine_config.azure_master_image.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			// Delete machine test
			{
				Config: func() string {
					var mcResource string
					if orchestrationApiVersion >= 120 {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "-AD", "ActiveDirectory")
					} else {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_without_serv_acc_uid_delete_machine, "-AD", "ActiveDirectory")
					}
					return composeTestResourceTf(
						mcResource,
						BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated),
						BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
						BuildHypervisorResourceAzure(t, hypervisor_testResources),
						BuildZoneResource(t, zoneInput, false),
					)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "name", name),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AD", "provisioning_scheme.identity_type", "ActiveDirectory"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestHybridAzureADMachineCatalogResourceAzure(t *testing.T) {
	ctx := context.Background()
	client := sharedClientForSweepers(ctx)

	orchestrationApiVersion := client.ClientConfig.OrchestrationApiVersion
	name := os.Getenv("TEST_MC_NAME") + "-HybAAD"
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: func() string {
					var mcResource string
					if orchestrationApiVersion >= 120 {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "-HybAAD", "HybridAzureAD")
					} else {
						mcResource = BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_without_serv_acc_uid, "-HybAAD", "HybridAzureAD")
					}
					return composeTestResourceTf(
						mcResource,
						BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated),
						BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
						BuildHypervisorResourceAzure(t, hypervisor_testResources),
						BuildZoneResource(t, zoneInput, false),
					)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "session_support", "MultiSession"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.identity_type", "HybridAzureAD"),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-HybAAD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			// Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "-HybAAD", "HybridAzureAD"),
					BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.azure_machine_config.azure_master_image.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.identity_type", "HybridAzureAD"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_AzureAd(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME"); v == "" {
		t.Fatal("TEST_MC_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_OFFERING"); v == "" {
		t.Fatal("TEST_MC_SERVICE_OFFERING must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_UPDATED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_IMAGE_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_IMAGE_STORAGE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_CONTAINER"); v == "" {
		t.Fatal("TEST_MC_IMAGE_CONTAINER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SUBNET"); v == "" {
		t.Fatal("TEST_MC_SUBNET must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_PROFILE_VM_NAME"); v == "" {
		t.Fatal("TEST_MC_MACHINE_PROFILE_VM_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_PROFILE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_PROFILE_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_VDA_AZURE_RESOURCE_GROUP must be set for acceptance tests")
	}
}

func TestAzureADMachineCatalogResourceAzure(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME") + "-AAD"
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_AzureAd(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAzureAd(t, machinecatalog_testResources_azure_ad),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "session_support", "MultiSession"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.identity_type", "AzureAD"),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-AAD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "delete_machine_accounts", "force_delete"},
			},
			// Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAzureAd(t, machinecatalog_testResources_azure_ad_updated),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.azure_machine_config.azure_master_image.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.identity_type", "AzureAD"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Workgroup(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME"); v == "" {
		t.Fatal("TEST_MC_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_OFFERING"); v == "" {
		t.Fatal("TEST_MC_SERVICE_OFFERING must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_UPDATED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_IMAGE_RESOURCE_GROUP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT"); v == "" {
		t.Fatal("TEST_MC_IMAGE_STORAGE_ACCOUNT must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_CONTAINER"); v == "" {
		t.Fatal("TEST_MC_IMAGE_CONTAINER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP"); v == "" {
		t.Fatal("TEST_MC_VDA_AZURE_RESOURCE_GROUP must be set for acceptance tests")
	}
}

func TestWorkgroupMachineCatalogResourceAzure(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME") + "-WRKGRP"
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Workgroup(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceWorkgroup(t, machinecatalog_testResources_workgroup),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "session_support", "MultiSession"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "provisioning_scheme.identity_type", "Workgroup"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-WG",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "delete_machine_accounts", "force_delete"},
			},
			// Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceWorkgroup(t, machinecatalog_testResources_workgroup_updated),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "provisioning_scheme.azure_machine_config.azure_master_image.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "provisioning_scheme.identity_type", "Workgroup"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_GCP(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_GCP"); v == "" {
		t.Fatal("TEST_MC_NAME_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_GCP"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_STORAGE_TYPE_GCP"); v == "" {
		t.Fatal("TEST_MC_STORAGE_TYPE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_AVAILABILITY_ZONES_GCP"); v == "" {
		t.Fatal("TEST_MC_AVAILABILITY_ZONES_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_PROFILE_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_PROFILE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_GCP"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_SNAPSHOT_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_SNAPSHOT_GCP must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_MC_DOMAIN_GCP"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_Subnet_GCP"); v == "" {
		t.Fatal("TEST_MC_Subnet_GCP must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceGCP(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_GCP")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_GCP")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_GCP(t)
			TestHypervisorResourcePoolPreCheck_GCP(t)
			TestMachineCatalogPreCheck_GCP(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceGCP(t, machinecatalog_testResources_gcp),
					BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp),
					BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.availability_zones", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceGCP(t, machinecatalog_testResources_gcp_updated),
					BuildHypervisorResourcePoolResourceGCP(t, hypervisor_resource_pool_testResource_gcp),
					BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Vsphere(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_NAME_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MEMORY_MB_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_MEMORY_MB_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CPU_COUNT_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_CPU_COUNT_VSPHERE must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceVsphere(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_VSPHERE")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_VSPHERE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Vsphere(t)
			TestHypervisorResourcePoolPreCheck_Vsphere(t)
			TestMachineCatalogPreCheck_Vsphere(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceVsphere(t, machine_catalog_testResources_vsphere),
					BuildHypervisorResourcePoolResourceVsphere(t, hypervisor_resource_pool_testResource_vsphere),
					BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_VSPHERE")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.vsphere_machine_config.master_image", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceVsphere(t, machine_catalog_testResources_vsphere_updated),
					BuildHypervisorResourcePoolResourceVsphere(t, hypervisor_resource_pool_testResource_vsphere),
					BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Xenserver(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_NAME_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MEMORY_MB_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_MEMORY_MB_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CPU_COUNT_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_CPU_COUNT_XENSERVER must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceXenserver(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_XENSERVER")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_XENSERVER")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Xenserver(t)
			TestHypervisorResourcePoolPreCheck_Xenserver(t)
			TestMachineCatalogPreCheck_Xenserver(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceXenserver(t, machine_catalog_testResources_xenserver),
					BuildHypervisorResourcePoolResourceXenServer(t, hypervisor_resource_pool_testResource_xenserver),
					BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_XENSERVER")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.xenserver_machine_config.master_image", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceXenserver(t, machine_catalog_testResources_xenserver_updated),
					BuildHypervisorResourcePoolResourceXenServer(t, hypervisor_resource_pool_testResource_xenserver),
					BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Nutanix(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_NAME_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CONTAINER_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_CONTAINER_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MEMORY_MB_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_MEMORY_MB_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CPU_COUNT_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_CPU_COUNT_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CORES_PER_CPU_COUNT_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_CORES_PER_CPU_COUNT_NUTANIX must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceNutanix(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_NUTANIX")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_NUTANIX")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Nutanix(t)
			TestHypervisorResourcePoolPreCheck_Nutanix(t)
			TestMachineCatalogPreCheck_Nutanix(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceNutanix(t, machine_catalog_testResources_nutanix),
					BuildHypervisorResourcePoolResourceNutanix(t, hypervisor_resource_pool_testResource_nutanix),
					BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.nutanix_machine_config.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_NUTANIX")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceNutanix(t, machine_catalog_testResources_nutanix_updated),
					BuildHypervisorResourcePoolResourceNutanix(t, hypervisor_resource_pool_testResource_nutanix),
					BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_SCVMM(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_SCVMM"); v == "" {
		t.Fatal("TEST_MC_NAME_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_SCVMM"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_SCVMM"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_SCVMM"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_SNAPSHOT_SCVMM"); v == "" {
		t.Fatal("TEST_MC_IMAGE_SNAPSHOT_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_NOTE_SCVMM"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_NOTE_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN_SCVMM"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MEMORY_MB_SCVMM"); v == "" {
		t.Fatal("TEST_MC_MEMORY_MB_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_CPU_COUNT_SCVMM"); v == "" {
		t.Fatal("TEST_MC_CPU_COUNT_SCVMM must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceSCVMM(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_SCVMM")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_SCVMM")
	masterImageNote := os.Getenv("TEST_MC_MASTER_IMAGE_NOTE_SCVMM")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_SCVMM(t)
			TestHypervisorResourcePoolPreCheck_SCVMM(t)
			TestMachineCatalogPreCheck_SCVMM(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceSCVMM(t, machine_catalog_testResources_scvmm),
					BuildHypervisorResourcePoolResourceSCVMM(t, hypervisor_resource_pool_testResource_scvmm),
					BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_SCVMM")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify master image note
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.scvmm_machine_config.master_image_note", masterImageNote),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceSCVMM(t, machine_catalog_testResources_scvmm_updated),
					BuildHypervisorResourcePoolResourceSCVMM(t, hypervisor_resource_pool_testResource_scvmm),
					BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
					// Verify updated master image note
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.scvmm_machine_config.master_image_note", masterImageNote+"-updated"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Aws_Ec2(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_NAME_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DOMAIN_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_DOMAIN_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_SERVICE_ACCOUNT_PASS_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_IMAGE_AMI_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_IMAGE_AMI_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MASTER_IMAGE_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_MASTER_IMAGE_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SERVICE_OFFERING_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_SERVICE_OFFERING_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_NETWORK_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_NETWORK_AWS_EC2 must be set for acceptance tests")
	}
}

func TestMachineCatalogResourceAwsEc2(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_AWS_EC2")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AWS_EC2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_AWS_EC2(t)
			TestHypervisorResourcePoolPreCheck_Aws_Ec2(t)
			TestMachineCatalogPreCheck_Aws_Ec2(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAwsEc2(t, machinecatalog_testResources_aws_ec2),
					BuildHypervisorResourcePoolResourceAwsEc2(t, hypervisor_resource_pool_testResource_aws_ec2),
					BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT_AWS_EC2")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify security groups
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.aws_machine_config.security_groups.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.aws_machine_config.image_ami", "provisioning_scheme.aws_machine_config.service_offering", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_domain_identity.service_account_password", "delete_machine_accounts", "force_delete"},
			},
			//Update description, master image and add machine test
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceAwsEc2(t, machinecatalog_testResources_aws_ec2_updated),
					BuildHypervisorResourcePoolResourceAwsEc2(t, hypervisor_resource_pool_testResource_aws_ec2),
					BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "Updated AWS EC2 MCS Machine Catalog"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_Azure(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_REGION_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_AZURE"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_AZURE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_AZURE"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_AZURE must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Azure(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_Azure(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedAzure(t, machinecatalog_testResources_manual_power_managed_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_GCP(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP"); v == "" {
		t.Fatal("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_GCP"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_GCP must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_GCP(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_GCP")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_GCP(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_GCP(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedGCP(t, machinecatalog_testResources_manual_power_managed_gcp),
					BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_Vsphere(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_DATACENTER_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_DATACENTER_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_HOST_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_HOST_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_VSPHERE"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Vsphere(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_VSPHERE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Vsphere(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_Vsphere(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedVsphere(t, machinecatalog_testResources_manual_power_managed_vsphere),
					BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_Xenserver(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_XENSERVER"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Xenserver(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_XENSERVER")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Xenserver(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_Xenserver(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedXenserver(t, machinecatalog_testResources_manual_power_managed_xenserver),
					BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_Nutanix(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_NUTANIX"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Nutanix(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_NUTANIX")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Nutanix(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_Nutanix(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedNutanix(t, machinecatalog_testResources_manual_power_managed_nutanix),
					BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_SCVMM(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_HOST_SCVMM"); v == "" {
		t.Fatal("TEST_MC_HOST_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_SCVMM"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_SCVMM"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_SCVMM(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_SCVMM")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_SCVMM(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_SCVMM(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedSCVMM(t, machinecatalog_testResources_manual_power_managed_scvmm),
					BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Power_Managed_AWS_EC2(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_MACHINE_NAME_MANUAL_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_AVAILABILITY_ZONE_MANUAL_AWS_EC2"); v == "" {
		t.Fatal("TEST_MC_AVAILABILITY_ZONE_MANUAL_AWS_EC2 must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Power_Managed_Aws_Ec2(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AWS_EC2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_AWS_EC2(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_AWS_EC2(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualPowerManagedAwsEc2(t, machinecatalog_testResources_manual_power_managed_aws_ec2),
					BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_Manual_Non_Power_Managed(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_MANUAL"); v == "" {
		t.Fatal("TEST_MC_NAME_MANUAL must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED"); v == "" {
		t.Fatal("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_Manual_Non_Power_Managed(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	zoneInput := os.Getenv("TEST_ZONE_INPUT")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestMachineCatalogPreCheck_Manual_Non_Power_Managed(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceManualNonPowerManaged(t, machinecatalog_testResources_manual_non_power_managed),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogNonManualPowerManaged", "name", name),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogNonManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogNonManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalogNonManualPowerManaged",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

func TestMachineCatalogPreCheck_RemotePC(t *testing.T) {
	if v := os.Getenv("TEST_MC_NAME_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_NAME_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_MACHINE_ACCOUNT_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_MACHINE_ACCOUNT_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_ALLOCATION_TYPE_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_ALLOCATION_TYPE_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_OU_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_OU_REMOTE_PC must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC"); v == "" {
		t.Fatal("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC must be set for acceptance tests")
	}
}

func TestMachineCatalogResource_RemotePC(t *testing.T) {
	name := os.Getenv("TEST_MC_NAME_REMOTE_PC")
	zoneInput := os.Getenv("TEST_ZONE_INPUT")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestMachineCatalogPreCheck_RemotePC(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildMachineCatalogResourceRemotePC(t, machinecatalog_testResources_remote_pc),
					BuildZoneResource(t, zoneInput, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", name),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalog",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed", "delete_machine_accounts", "force_delete"},
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

var (
	machinecatalog_testResources_zeroMachines_azure = `
	resource "citrix_machine_catalog" "testMachineCatalogMachines" {
		name                		= "ZeroMachinesCatalog"
		description					= "Machine Catalog with zero machines"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				vda_resource_group = "%s"
			}
			number_of_total_machines = 	0
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_zeroMachines_azureUpdated = `
	resource "citrix_machine_catalog" "testMachineCatalogMachines" {
		name                		= "UpdatedOneMachineCatalog"
		description					= "Machine Catalog with updated Machine Count"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "ActiveDirectory"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				vda_resource_group = "%s"
			}
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_azure = `
resource "citrix_machine_catalog" "testMachineCatalog%s" {
	name                		= "%s"
	description					= "on prem catalog for import testing"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "MCS"
	minimum_functional_level    = "L7_9"
	provisioning_scheme			= 	{
		hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
		hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
		identity_type = "%s"
		machine_domain_identity = {
			domain 						= "%s"
			service_account_id 			= citrix_service_account.testServiceAccountAD.id
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			azure_master_image 	 = {
				resource_group 		 = "%s"
				storage_account 	 = "%s"
				container 			 = "%s"
				master_image		 = "%s"
			}
			storage_type = "StandardSSD_LRS"
			use_managed_disks = true
			writeback_cache = {
				wbc_disk_storage_type = "StandardSSD_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
				writeback_cache_memory_size_mb = 256
				storage_cost_saving = true
			}
			vda_resource_group = "%s"
		}
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "%s"
			naming_scheme_type ="Numeric"
		}
	}

	delete_machine_accounts 	= "Delete"
	zone						= citrix_zone.test.id
}
`
	machinecatalog_testResources_azure_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		minimum_functional_level    = "L7_20"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			availability_zones = ["1","3"]
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`

	machinecatalog_testResources_azure_delete_machine = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"		
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 	 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account_id 			= citrix_service_account.testServiceAccountAD.id
			}
			azure_machine_config = {
				service_offering 	 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			availability_zones = ["1","3"]
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_azure_without_serv_acc_uid = `
resource "citrix_machine_catalog" "testMachineCatalog%s" {
	name                		= "%s"
	description					= "on prem catalog for import testing"
	allocation_type				= "Random"
	session_support				= "MultiSession"
	provisioning_type			= "MCS"
	minimum_functional_level    = "L7_9"
	provisioning_scheme			= 	{
		hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
		hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
		identity_type = "%s"
		machine_domain_identity = {
			domain 						= "%s"
			service_account				= "%s"
			service_account_password 	= "%s"
		}
		azure_machine_config = {
			service_offering 	 = "%s"
			azure_master_image 	 = {
				resource_group 		 = "%s"
				storage_account 	 = "%s"
				container 			 = "%s"
				master_image		 = "%s"
			}
			storage_type = "StandardSSD_LRS"
			use_managed_disks = true
			writeback_cache = {
				wbc_disk_storage_type = "StandardSSD_LRS"
				persist_wbc = true
				persist_os_disk = true
				persist_vm = true
				writeback_cache_disk_size_gb = 127
				writeback_cache_memory_size_mb = 256
				storage_cost_saving = true
			}
		}
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		number_of_total_machines = 	1
		machine_account_creation_rules ={
			naming_scheme =     "%s"
			naming_scheme_type ="Numeric"
		}
	}
	delete_machine_accounts 	= "Delete"
	zone						= citrix_zone.test.id
}
`
	machinecatalog_testResources_azure_without_serv_acc_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		minimum_functional_level    = "L7_20"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}
		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}`
	machinecatalog_testResources_azure_without_serv_acc_uid_delete_machine = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"		
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 	 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			azure_machine_config = {
				service_offering 	 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			availability_zones = ["1","3"]
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}
		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_azure_ad = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "on prem catalog for import testing"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "AzureAD"
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				machine_profile = {
					machine_profile_vm_name = "%s"
					machine_profile_resource_group = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_azure_ad_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "AzureAD"
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				machine_profile = {
					machine_profile_vm_name = "%s"
					machine_profile_resource_group = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			network_mapping = [
				{
					network_device = "0"
					network 	   = "%s"
				}
			]
			availability_zones = ["1","3"]
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`

	machinecatalog_testResources_workgroup = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "on prem catalog for import testing"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "Workgroup"
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			number_of_total_machines = 	1
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}

		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`
	machinecatalog_testResources_workgroup_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog%s" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_azure_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "Workgroup"
			azure_machine_config = {
				service_offering 	 = "%s"
				azure_master_image 	 = {
					resource_group 		 = "%s"
					storage_account 	 = "%s"
					container 			 = "%s"
					master_image		 = "%s"
				}
				storage_type = "StandardSSD_LRS"
				use_managed_disks = true
				writeback_cache = {
					wbc_disk_storage_type = "StandardSSD_LRS"
					persist_wbc = true
					persist_os_disk = true
					persist_vm = true
					writeback_cache_disk_size_gb = 127
					writeback_cache_memory_size_mb = 256
					storage_cost_saving = true
				}
				vda_resource_group = "%s"
			}
			availability_zones = ["1","3"]
			number_of_total_machines = 	2
			machine_account_creation_rules ={
				naming_scheme =     "%s"
				naming_scheme_type ="Numeric"
			}
		}
		
		delete_machine_accounts 	= "Delete"
		zone						= citrix_zone.test.id
	}
	`

	machinecatalog_testResources_gcp = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "on prem catalog for import testing"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		minimum_functional_level    = "L7_9"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_gcp_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			gcp_machine_config = {
				storage_type = "%s"
				machine_profile = "%s"
				master_image		 = "%s"
				machine_snapshot = "%s"
			}
			number_of_total_machines = 	1
			availability_zones = %s
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_zone.test.id
	}
	`

	machinecatalog_testResources_gcp_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "updatedCatalog"
		allocation_type				= "Random"
		session_support				= "MultiSession"
		provisioning_type			= "MCS"
		minimum_functional_level    = "L7_20"
		provisioning_scheme			= 	{
			hypervisor			 = citrix_gcp_hypervisor.testHypervisor.id
			hypervisor_resource_pool = citrix_gcp_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type = "%s"
			machine_domain_identity = {
				domain 						= "%s"
				service_account				= "%s"
				service_account_password 	= "%s"
			}
			gcp_machine_config = {
				storage_type = "%s"
				machine_profile = "%s"
				master_image		 = "%s"
				machine_snapshot = "%s"
			}
			number_of_total_machines = 	2
			availability_zones = %s
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
		zone						= citrix_zone.test.id
	}
	`
	machine_catalog_testResources_vsphere = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "vsphere catalog for acceptance testing"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 1
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_vsphere_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    vsphere_machine_config = {
    	        master_image_vm = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_vsphere_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "updatedCatalog"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 2
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_vsphere_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_vsphere_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    vsphere_machine_config = {
    	        master_image_vm = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_scvmm = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "scvmm catalog for acceptance testing"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 1
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_scvmm_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    scvmm_machine_config = {
    	        master_image = "%s"
				image_snapshot = "%s"
				master_image_note = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_scvmm_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "updatedCatalog"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 2
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_scvmm_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_scvmm_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    scvmm_machine_config = {
    	        master_image = "%s"
				image_snapshot = "%s"
				master_image_note = "%s-updated"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_xenserver = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "xenserver catalog for acceptance testing"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 1
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    xenserver_machine_config = {
    	        master_image_vm = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_xenserver_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "updatedCatalog"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 2
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_xenserver_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    xenserver_machine_config = {
    	        master_image_vm = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_nutanix = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "nutanix catalog for acceptance testing"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 1
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_nutanix_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    nutanix_machine_config = {
				container = "%s"
    	        master_image = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
				cores_per_cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machine_catalog_testResources_nutanix_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
    	description                 = "updatedCatalog"
    	provisioning_type = "MCS"
    	allocation_type             = "Random"
    	session_support             = "MultiSession"
    	zone                        = citrix_zone.test.id
    	provisioning_scheme         = {
    	    identity_type = "ActiveDirectory"
    	    number_of_total_machines = 2
    	    machine_account_creation_rules = {
    	        naming_scheme = "test-machine-##"
    	        naming_scheme_type = "Numeric"
    	    }
    	    hypervisor = citrix_nutanix_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_nutanix_hypervisor_resource_pool.testHypervisorResourcePool.id
    	    nutanix_machine_config = {
				container = "%s"
    	        master_image = "%s"
    	        memory_mb = "%s"
				cpu_count = "%s"
				cores_per_cpu_count = "%s"
    	    }
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
    	}
	}
	`

	machinecatalog_testResources_aws_ec2 = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
		description                 = "AWS EC2 MCS Machine Catalog"
		zone                        = citrix_zone.test.id
		allocation_type             = "Random"
		session_support             = "MultiSession"
		provisioning_type           = "MCS"
		provisioning_scheme         =   {
    	    hypervisor = citrix_aws_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type      = "ActiveDirectory"
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
			aws_machine_config = {
				image_ami = "%s"
				master_image = "%s"
				service_offering = "%s"
				security_groups = [
					"default"
				]
				tenancy_type = "Shared"
			}
			network_mapping = [
				{
					network_device = "0"
					network = "%s"
				}
			]
			number_of_total_machines =  1
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
	}
	`

	machinecatalog_testResources_aws_ec2_updated = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                        = "%s"
		description                 = "Updated AWS EC2 MCS Machine Catalog"
		zone                        = citrix_zone.test.id
		allocation_type             = "Random"
		session_support             = "MultiSession"
		provisioning_type           = "MCS"
		provisioning_scheme         =   {
    	    hypervisor = citrix_aws_hypervisor.testHypervisor.id
    	    hypervisor_resource_pool = citrix_aws_hypervisor_resource_pool.testHypervisorResourcePool.id
			identity_type      = "ActiveDirectory"
    	    machine_domain_identity = {
    	        service_account             = "%s"
			    domain = "%s"
    	        service_account_password    = "%s"
    	    }
			aws_machine_config = {
				image_ami = "%s"
				master_image = "%s"
				service_offering = "%s"
				security_groups = [
					"default"
				]
				tenancy_type = "Shared"
			}
			network_mapping = [
				{
					network_device = "0"
					network = "%s"
				}
			]
			number_of_total_machines =  2
			machine_account_creation_rules ={
				naming_scheme =     "test-machine-##"
				naming_scheme_type ="Numeric"
			}
		}
	}
	`

	machinecatalog_testResources_manual_power_managed_azure = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                		= "%s"
		description					= "manual power managed multi-session catalog testing for Azure Hypervisor"
		zone						= citrix_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= true
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				hypervisor = citrix_azure_hypervisor.testHypervisor.id
				machines = [
					{
						region = "%s"
						resource_group_name = "%s"
						machine_name = "%s"
						machine_account = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_gcp = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                		= "%s"
		description					= "manual power managed multi-session catalog testing"
		zone						= citrix_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= true
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				hypervisor = citrix_gcp_hypervisor.testHypervisor.id
				machines = [
					{
						region = "%s"
						project_name = "%s"
						machine_name = "%s"
						machine_account = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_vsphere = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                        = "%s"
		description                 = "manual power managed multi-session catalog testing"
		is_power_managed = true
		is_remote_pc = false
		provisioning_type = "Manual"
		allocation_type             = "%s"
		session_support             = "%s"
		zone                        = citrix_zone.test.id
		machine_accounts = [
			{
				hypervisor = citrix_vsphere_hypervisor.testHypervisor.id
				machines = [
					{
						datacenter = "%s"
						host = "%s"
						machine_name = "%s"
						machine_account = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_xenserver = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                        = "%s"
		description                 = "manual power managed multi-session catalog testing"
		is_power_managed 			= true
		is_remote_pc 				= false
		provisioning_type 			= "Manual"
		allocation_type             = "%s"
		session_support             = "%s"
		zone                        = citrix_zone.test.id
		machine_accounts = [
			{
				hypervisor = citrix_xenserver_hypervisor.testHypervisor.id
				machines = [
					{
						machine_name = "%s"
						machine_account = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_nutanix = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                        = "%s"
		description                 = "manual power managed multi-session catalog testing"
		is_power_managed 			= true
		is_remote_pc 				= false
		provisioning_type 			= "Manual"
		allocation_type             = "%s"
		session_support             = "%s"
		zone                        = citrix_zone.test.id
		machine_accounts = [
			{
				hypervisor = citrix_nutanix_hypervisor.testHypervisor.id
				machines = [
					{
						machine_name = "%s"
						machine_account = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_scvmm = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                        = "%s"
		description                 = "manual power managed multi-session catalog testing"
		is_power_managed 			= true
		is_remote_pc 				= false
		provisioning_type 			= "Manual"
		allocation_type             = "%s"
		session_support             = "%s"
		zone                        = citrix_zone.test.id
		machine_accounts = [
			{
				hypervisor = citrix_scvmm_hypervisor.testHypervisor.id
				machines = [
					{
						machine_name = "%s"
						machine_account = "%s"
						host = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_power_managed_aws_ec2 = `
	resource "citrix_machine_catalog" "testMachineCatalogManualPowerManaged" {
		name                        = "%s"
		description                 = "manual power managed multi-session catalog testing"
		is_power_managed 			= true
		is_remote_pc 				= false
		provisioning_type 			= "Manual"
		allocation_type             = "%s"
		session_support             = "%s"
		zone                        = citrix_zone.test.id
		machine_accounts = [
			{
				hypervisor = citrix_aws_hypervisor.testHypervisor.id
				machines = [
					{
						machine_name = "%s"
						machine_account = "%s"
                    	availability_zone = "%s"
					}
				]
			}
		]
	}
	`

	machinecatalog_testResources_manual_non_power_managed = `
	resource "citrix_machine_catalog" "testMachineCatalogNonManualPowerManaged" {
		name                		= "%s"
		description					= "manual non power managed multi-session catalog testing"
		zone						= citrix_zone.test.id
		allocation_type				= "%s"
		session_support				= "%s"
		is_power_managed			= false
		is_remote_pc			    = false
		provisioning_type			= "Manual"
		machine_accounts = [
			{
				machines = [
					{
						machine_account = "%s"
					}
				]
			}
		]
	}
	`
	machinecatalog_testResources_remote_pc = `
	resource "citrix_machine_catalog" "testMachineCatalog" {
		name                		= "%s"
		description					= "on prem catalog for import testing remotePC"
		allocation_type				= "%s"
		session_support				= "SingleSession"
		provisioning_type			= "Manual"
		is_remote_pc				= true
		is_power_managed			= false
		machine_accounts			= [
			{
				machines = [
					{
						machine_account = "%s"
					}
				]
			}
		]
		remote_pc_ous = [
			{
				include_subfolders = %s
				ou_name = "%s"
			}
		]
		zone						= citrix_zone.test.id
	}
	`
)

func BuildMachineCatalogResourceAzure(t *testing.T, machineResource, catalogNameSuffix, identityType string) string {
	name := os.Getenv("TEST_MC_NAME")
	namingScheme := "vda-##"
	if identityType == "HybridAzureAD" {
		name += "-HybAAD"
		namingScheme += "-HybAAD"
	}
	if identityType == "AzureAD" {
		name += "-AAD"
		namingScheme += "-AAD"
	}
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS")
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	subnet := os.Getenv("TEST_MC_SUBNET")
	vdaResourceGroup := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP")
	if machineResource == machinecatalog_testResources_azure_updated {
		master_image = os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")
	}

	//machine account
	domain := os.Getenv("TEST_MC_DOMAIN")

	switch machineResource {
	case machinecatalog_testResources_azure, machinecatalog_testResources_azure_delete_machine:
		// Do not pass service account id and password since we are using service account resource
		return fmt.Sprintf(machineResource, catalogNameSuffix, name, identityType, domain, service_offering, resource_group, storage_account, container, master_image, vdaResourceGroup, subnet, namingScheme)
	case machinecatalog_testResources_azure_without_serv_acc_uid, machinecatalog_testResources_azure_without_serv_acc_uid_delete_machine, machinecatalog_testResources_azure_without_serv_acc_updated:
		// Do not pass service account id and password since we are using service account resource
		return fmt.Sprintf(machineResource, catalogNameSuffix, name, identityType, domain, service_account, service_account_pass, service_offering, resource_group, storage_account, container, master_image, subnet, namingScheme)
	}

	return fmt.Sprintf(machineResource, catalogNameSuffix, name, identityType, domain, service_account, service_account_pass, service_offering, resource_group, storage_account, container, master_image, vdaResourceGroup, subnet, namingScheme)
}

func BuildMachineCatalogResourceAzureAd(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME") + "-AAD"
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	subnet := os.Getenv("TEST_MC_SUBNET")
	vdaResourceGroup := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP")
	namingScheme := "vda-##-AAD"

	machine_profile_vm_name := os.Getenv("TEST_MC_MACHINE_PROFILE_VM_NAME")
	machine_profile_resource_group := os.Getenv("TEST_MC_MACHINE_PROFILE_RESOURCE_GROUP")
	if machineResource == machinecatalog_testResources_azure_ad_updated {
		master_image = os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")
	}

	return fmt.Sprintf(machineResource, "-AAD", name, service_offering, resource_group, storage_account, container, master_image, machine_profile_vm_name, machine_profile_resource_group, vdaResourceGroup, subnet, namingScheme)
}

func BuildMachineCatalogResourceWorkgroup(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME") + "-WRKGRP"
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	vdaResourceGroup := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP")
	namingScheme := "vda-##-WG"

	if machineResource == machinecatalog_testResources_workgroup_updated {
		master_image = os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")
	}

	return fmt.Sprintf(machineResource, "-WG", name, service_offering, resource_group, storage_account, container, master_image, vdaResourceGroup, namingScheme)
}

func BuildMachineCatalogResourceGCP(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_GCP")
	identityType := "ActiveDirectory"
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_GCP")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_GCP")
	storage_type := os.Getenv("TEST_MC_STORAGE_TYPE_GCP")
	availability_zones_list := strings.Split(os.Getenv("TEST_MC_AVAILABILITY_ZONES_GCP"), ",")
	availability_zones := "[\"" + strings.Join(availability_zones_list, "\",\"") + "\"]" // ["1","3"]
	machine_profile := os.Getenv("TEST_MC_MACHINE_PROFILE_GCP")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_GCP")
	machine_snapshot := os.Getenv("TEST_MC_MACHINE_SNAPSHOT_GCP")

	//machine account
	domain := os.Getenv("TEST_MC_DOMAIN_GCP")

	return fmt.Sprintf(machineResource, name, identityType, domain, service_account, service_account_pass, storage_type, machine_profile, master_image, machine_snapshot, availability_zones)
}

func BuildMachineCatalogResourceVsphere(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_VSPHERE")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_VSPHERE")
	memory_mb := os.Getenv("TEST_MC_MEMORY_MB_VSPHERE")
	cpu_count := os.Getenv("TEST_MC_CPU_COUNT_VSPHERE")
	domain := os.Getenv("TEST_MC_DOMAIN_VSPHERE")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_VSPHERE")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_VSPHERE")

	return fmt.Sprintf(machineResource, name, master_image, memory_mb, cpu_count, service_account, domain, service_account_pass)
}

func BuildMachineCatalogResourceSCVMM(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_SCVMM")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_SCVMM")
	image_snapshot := os.Getenv("TEST_MC_IMAGE_SNAPSHOT_SCVMM")
	master_image_note := os.Getenv("TEST_MC_MASTER_IMAGE_NOTE_SCVMM")
	memory_mb := os.Getenv("TEST_MC_MEMORY_MB_SCVMM")
	cpu_count := os.Getenv("TEST_MC_CPU_COUNT_SCVMM")
	domain := os.Getenv("TEST_MC_DOMAIN_SCVMM")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_SCVMM")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_SCVMM")

	return fmt.Sprintf(machineResource, name, master_image, image_snapshot, master_image_note, memory_mb, cpu_count, service_account, domain, service_account_pass)
}

func BuildMachineCatalogResourceXenserver(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_XENSERVER")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_XENSERVER")
	memory_mb := os.Getenv("TEST_MC_MEMORY_MB_XENSERVER")
	cpu_count := os.Getenv("TEST_MC_CPU_COUNT_XENSERVER")
	domain := os.Getenv("TEST_MC_DOMAIN_XENSERVER")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_XENSERVER")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_XENSERVER")

	return fmt.Sprintf(machineResource, name, master_image, memory_mb, cpu_count, service_account, domain, service_account_pass)
}

func BuildMachineCatalogResourceNutanix(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_NUTANIX")
	container := os.Getenv("TEST_MC_CONTAINER_NUTANIX")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_NUTANIX")
	memory_mb := os.Getenv("TEST_MC_MEMORY_MB_NUTANIX")
	cpu_count := os.Getenv("TEST_MC_CPU_COUNT_NUTANIX")
	cores_per_cpu_count := os.Getenv("TEST_MC_CORES_PER_CPU_COUNT_NUTANIX")
	domain := os.Getenv("TEST_MC_DOMAIN_NUTANIX")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_NUTANIX")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_NUTANIX")

	return fmt.Sprintf(machineResource, name, container, master_image, memory_mb, cpu_count, cores_per_cpu_count, service_account, domain, service_account_pass)
}

func BuildMachineCatalogResourceAwsEc2(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_AWS_EC2")
	domain := os.Getenv("TEST_MC_DOMAIN_AWS_EC2")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT_AWS_EC2")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS_AWS_EC2")
	image_ami := os.Getenv("TEST_MC_IMAGE_AMI_AWS_EC2")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE_AWS_EC2")
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING_AWS_EC2")
	network := os.Getenv("TEST_MC_NETWORK_AWS_EC2")

	return fmt.Sprintf(machineResource, name, service_account, domain, service_account_pass, image_ami, master_image, service_offering, network)
}

func BuildMachineCatalogWithZeroMachines(t *testing.T, machineResource string) string {
	domain := os.Getenv("TEST_MC_DOMAIN")
	service_account := os.Getenv("TEST_MC_SERVICE_ACCOUNT")
	service_account_pass := os.Getenv("TEST_MC_SERVICE_ACCOUNT_PASS")
	service_offering := os.Getenv("TEST_MC_SERVICE_OFFERING")
	resource_group := os.Getenv("TEST_MC_IMAGE_RESOURCE_GROUP")
	storage_account := os.Getenv("TEST_MC_IMAGE_STORAGE_ACCOUNT")
	container := os.Getenv("TEST_MC_IMAGE_CONTAINER")
	master_image := os.Getenv("TEST_MC_MASTER_IMAGE")
	vdaResourceGroup := os.Getenv("TEST_MC_VDA_AZURE_RESOURCE_GROUP")

	return fmt.Sprintf(machineResource, domain, service_account, service_account_pass, service_offering, resource_group, storage_account, container, master_image, vdaResourceGroup)
}

func BuildMachineCatalogResourceManualPowerManagedAzure(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_AZURE")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_AZURE")
	region := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED")
	resource_group := os.Getenv("TEST_MC_RESOURCE_GROUP_MANUAL_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, region, resource_group, machine_name, machine_account)
}

func BuildMachineCatalogResourceManualPowerManagedGCP(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_GCP")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_GCP")
	region := os.Getenv("TEST_MC_REGION_MANUAL_POWER_MANAGED_GCP")
	project_name := os.Getenv("TEST_MC_PROJECT_NAME_MANUAL_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, region, project_name, machine_name, machine_account)
}

func BuildMachineCatalogResourceManualPowerManagedVsphere(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	datacenter := os.Getenv("TEST_MC_DATACENTER_VSPHERE")
	host := os.Getenv("TEST_MC_HOST_VSPHERE")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_VSPHERE")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_VSPHERE")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, datacenter, host, machine_name, machine_account)
}

func BuildMachineCatalogResourceManualPowerManagedXenserver(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_XENSERVER")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_XENSERVER")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_name, machine_account)
}

func BuildMachineCatalogResourceManualPowerManagedNutanix(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_NUTANIX")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_NUTANIX")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_name, machine_account)
}

func BuildMachineCatalogResourceManualPowerManagedSCVMM(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_SCVMM")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_SCVMM")
	host := os.Getenv("TEST_MC_HOST_SCVMM")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_name, machine_account, host)
}

func BuildMachineCatalogResourceManualPowerManagedAwsEc2(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")
	machine_name := os.Getenv("TEST_MC_MACHINE_NAME_MANUAL_AWS_EC2")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_AWS_EC2")
	availability_zone := os.Getenv("TEST_MC_AVAILABILITY_ZONE_MANUAL_AWS_EC2")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_name, machine_account, availability_zone)
}

func BuildMachineCatalogResourceManualNonPowerManaged(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_MANUAL")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_MANUAL_NON_POWER_MANAGED")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_MANUAL_NON_POWER_MANAGED")
	session_support := os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_NON_POWER_MANAGED")

	return fmt.Sprintf(machineResource, name, allocation_type, session_support, machine_account)
}

func BuildMachineCatalogResourceRemotePC(t *testing.T, machineResource string) string {
	name := os.Getenv("TEST_MC_NAME_REMOTE_PC")
	machine_account := os.Getenv("TEST_MC_MACHINE_ACCOUNT_REMOTE_PC")
	allocation_type := os.Getenv("TEST_MC_ALLOCATION_TYPE_REMOTE_PC")
	ou := os.Getenv("TEST_MC_OU_REMOTE_PC")
	include_subfolders := os.Getenv("TEST_MC_INCLUDE_SUBFOLDERS_REMOTE_PC")

	return fmt.Sprintf(machineResource, name, allocation_type, machine_account, include_subfolders, ou)
}

func machineCatalogSweeper(ctx context.Context, machineCatalogName string, client *citrixclient.CitrixDaasClient) error {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogName)
	machineCatalog, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting machine catalog: %s", err)
	}
	deleteMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsDeleteMachineCatalog(ctx, machineCatalog.GetId())
	httpResp, err = citrixclient.AddRequestData(deleteMachineCatalogRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", machineCatalogName, err)
	}
	return nil
}
