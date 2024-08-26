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

// testHypervisorPreCheck validates the necessary env variable exist
// in the testing environment
func TestAzureMcsSuitePreCheck(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	TestProviderPreCheck(t)
	TestHypervisorPreCheck_Azure(t)
	TestHypervisorResourcePoolPreCheck_Azure(t)
	TestMachineCatalogPreCheck_Azure(t)
	TestMachineCatalogPreCheck_Manual_Power_Managed_Azure(t)
	TestDeliveryGroupPreCheck(t)
	TestAdminFolderPreCheck(t)
	TestApplicationResourcePreCheck(t)
	TestAdminScopeResourcePreCheck(t)
	TestAdminRolePreCheck(t)
	TestPolicySetResourcePreCheck(t)

	if !isOnPremises {
		TestMachineCatalogPreCheck_AzureAd(t)
		TestMachineCatalogPreCheck_Workgroup(t)
	} else {
		TestAdminUserPreCheck(t)
	}
}

func TestAzureMcs(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}
	gotestContext := os.Getenv("GOTEST_CONTEXT")
	isGitHubAction := false
	if gotestContext != "" && gotestContext == "github" {
		// Tests being run in GitHub Action
		isGitHubAction = true
	}

	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")
	zoneDescription := os.Getenv("TEST_ZONE_DESCRIPTION")
	if zoneInput == "" {
		zoneInput = "second zone"
		zoneDescription = "description for go test zone"
	}

	hypervName := os.Getenv("TEST_HYPERV_NAME_AZURE")
	resourcePoolName := os.Getenv("TEST_HYPERV_RP_NAME")
	machineCatalogName := os.Getenv("TEST_MC_NAME")
	deliveryGroupName := os.Getenv("TEST_DG_NAME")
	hybridCatalogName := machineCatalogName + "-HybAAD"
	aadCatalogName := machineCatalogName + "-AAD"
	workgroupCatalogName := machineCatalogName + "-WRKGRP"
	manualPmCatalogName := os.Getenv("TEST_MC_NAME_MANUAL")
	applicationName := os.Getenv("TEST_APP_NAME")
	appFolderName := os.Getenv("TEST_ADMIN_FOLDER_NAME")
	folder_name_1 := fmt.Sprintf("%s-1", appFolderName)
	folder_name_1_updated := fmt.Sprintf("%s-1-updated", appFolderName)
	folder_name_2 := fmt.Sprintf("%s-2", appFolderName)
	adminScopeName := os.Getenv("TEST_ADMIN_SCOPE_NAME")
	adminRoleName := os.Getenv("TEST_ROLE_NAME")
	userName := os.Getenv("TEST_ADMIN_USER_NAME")
	userDomainName := os.Getenv("TEST_ADMIN_USER_DOMAIN")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Manual_Power_Managed_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestAdminFolderPreCheck(t)
			TestApplicationResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			/****************** Zone Test ******************/
			// Zone - Create and Read testing
			{
				Config: BuildZoneResource(t, zoneInput, false),
				Check:  getAggregateTestFunc(isOnPremises, zoneInput, zoneDescription),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "metadata"},
			},
			// Update and Read testing
			{
				Config: BuildZoneResource(t, zoneInput, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "name", fmt.Sprintf("%s-updated", zoneInput)),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "description", fmt.Sprintf("updated %s", zoneDescription)),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.#", "4"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.3.name", "key4"),
					resource.TestCheckResourceAttr("citrix_zone.test", "metadata.3.value", "value4"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},

			/****************** Hypervisor Test ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, true),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_azure_hypervisor.testHypervisor", "name", hypervName),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_azure_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "application_secret"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_azure_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", hypervName)),
				),
			},

			/****************** Resource Pool Test ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "name", resourcePoolName),
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
			},
			// Update and Read
			{
				Config: composeTestResourceTf(
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_azure_hypervisor_resource_pool.testHypervisorResourcePool", "name", fmt.Sprintf("%s-updated", resourcePoolName)),
				),
			},

			/****************** Machine Catalog Test - MCS AD / AAD / HybridAAD - Manual Power Managed ******************/
			// Create and Read testing
			{
				Config: composeAzureMachineCatalogTestResourceTf(t, isOnPremises),
				Check: resource.ComposeAggregateTestCheckFunc(
					/*** Verify MCS AD Machine Catalog ***/
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", machineCatalogName),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "session_support", "MultiSession"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),

					/*** Verify MCS Hybrid AAD Machine Catalog ***/
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "name", hybridCatalogName),
					// Verify domain FQDN
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "session_support", "MultiSession"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.identity_type", "HybridAzureAD"),
					// Verify domain admin username
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.machine_domain_identity.service_account", os.Getenv("TEST_MC_SERVICE_ACCOUNT")),
					// Verify nic network
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-HybAAD", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),

					/*** Optional - Verify MCS AAD / WorkGroup Machine Catalog ***/
					composeCloudMachineCatalogTestVerification(isOnPremises, aadCatalogName, workgroupCatalogName),

					/*** Verify Manual Power Managed Machine Catalog ***/
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "name", manualPmCatalogName),
					// Verify session support
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "session_support", os.Getenv("TEST_MC_SESSION_SUPPORT_MANUAL_POWER_MANAGED")),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalogManualPowerManaged", "machine_accounts.#", "1"),
				),
			},
			// ImportState testing - MCS AD
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password"},
				SkipFunc:                skipForGitHubAction(isGitHubAction),
			},
			// ImportState testing - MCS HybridAAD
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-HybAAD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache", "provisioning_scheme.machine_domain_identity.service_account", "provisioning_scheme.machine_config.service_account_password"},
				SkipFunc:                skipForGitHubAction(isGitHubAction),
			},
			// ImportState testing - MCS AAD
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-AAD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache"},
				SkipFunc:                func() (bool, error) { return (isOnPremises || isGitHubAction), nil },
			},
			// ImportState testing
			{
				ResourceName:      "citrix_machine_catalog.testMachineCatalog-WG",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"provisioning_scheme.network_mapping", "provisioning_scheme.azure_machine_config.writeback_cache"},
				SkipFunc:                func() (bool, error) { return (isOnPremises || isGitHubAction), nil },
			},
			// ImportState testing - Manual Power Managed
			{
				ResourceName:            "citrix_machine_catalog.testMachineCatalogManualPowerManaged",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"machine_accounts", "is_remote_pc", "is_power_managed"},
			},

			/****************** Delivery Group Test - Create ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "name", deliveryGroupName),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "desktops.#", "2"),
					// Verify number of reboot schedules
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "reboot_schedules.#", "2"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "1"),
					// Verify the policy set id is not assigned to the delivery group
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_delivery_group.testDeliveryGroup",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "autoscale_settings", "associated_machine_catalogs", "reboot_schedules"},
				SkipFunc:                skipForGitHubAction(isGitHubAction),
			},

			/****************** Delivery Group Test - MCS AD - Update ******************/
			// Machine Catalog: Update description, master image and add machine test
			// Delivery Group: Update name, description and add machine testing
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updated),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					/*** Verify Machine Catalog ***/
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", machineCatalogName),
					// Verify updated description
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "description", "updatedCatalog"),
					// Verify updated image
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.azure_machine_config.azure_master_image.master_image", os.Getenv("TEST_MC_MASTER_IMAGE_UPDATED")),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "2"),

					/*** Verify Delivery Group ***/
					// Verify name of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "name", fmt.Sprintf("%s-updated", deliveryGroupName)),
					// Verify description of delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "description", "Delivery Group for testing updated"),
					// Verify number of desktops
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "desktops.#", "1"),
					// Verify number of reboot schedules
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "reboot_schedules.#", "1"),
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "2"),
					// Verify the policy set id is not assigned to the delivery group
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
			},

			/****************** Delivery Group Test - Policy Update ******************/
			// Create Policy and assign to delivery group
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources_updatedWithPolicySetId),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the policy set id assigned to the delivery group
					resource.TestCheckResourceAttrSet("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
				),
				SkipFunc: func() (bool, error) { return true, nil }, // Always skip this test on testing
			},

			/****************** Machine Catalog & Delivery Group Test - MCS AD - Delete Machine ******************/
			// Delivery Group: Remove machine test
			// Machine Catalog: Delete machine test
			{
				Config: composeTestResourceTf(
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify total number of machines in delivery group
					resource.TestCheckResourceAttr("citrix_delivery_group.testDeliveryGroup", "total_machines", "1"),
					// Verify the policy set id assigned to the delivery group is removed
					resource.TestCheckNoResourceAttr("citrix_delivery_group.testDeliveryGroup", "policy_set_id"),
					// Verify updated name of catalog
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "name", machineCatalogName),
					// Verify total number of machines
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.number_of_total_machines", "1"),
					// Verify machine catalog identity type
					resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog", "provisioning_scheme.identity_type", "ActiveDirectory"),
				),
			},

			/****************** Admin Folder Test ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminFolderResource(t, testAdminFolderResource, "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_admin_folder.testAdminFolder2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update type testing
			{
				Config: composeTestResourceTf(
					BuildAdminFolderResource(t, testAdminFolderResource, "ContainsApplicationGroups"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplicationGroups"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplicationGroups"),
				),
			},
			// Update name and parent path testing
			{
				Config: composeTestResourceTf(
					BuildAdminFolderResource(t, testAdminFolderResource_nameAndParentPathUpdated1, "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify parent path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "parent_path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\%s\\", folder_name_1_updated, folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// Update name and remove parent path testing
			{
				Config: composeTestResourceTf(
					BuildAdminFolderResource(t, testAdminFolderResource_parentPathRemoved, "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},
			// Update name and remove parent path testing
			{
				Config: composeTestResourceTf(
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "name", folder_name_1_updated),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "name", folder_name_2),
					// Verify path of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "path", fmt.Sprintf("%s\\", folder_name_1_updated)),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "path", fmt.Sprintf("%s\\", folder_name_2)),
					// Verify type of admin folder
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder1", "type.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsMachineCatalogs"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder1", "type.*", "ContainsApplications"),
					resource.TestCheckResourceAttr("citrix_admin_folder.testAdminFolder2", "type.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsMachineCatalogs"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_folder.testAdminFolder2", "type.*", "ContainsApplications"),
				),
			},

			/****************** Application Test / Policy Test / Admin Scope / Admin Role Test - In Parallel ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminRoleResource(t, adminRoleTestResource),
					BuildAdminScopeResource(t, adminScopeTestResource),
					// BuildPolicySetResource(t, policy_set_testResource),
					BuildApplicationResource(t, testApplicationResource),
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					/*** Verify Application ***/
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "name", applicationName),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "description", "Application for testing"),
					// Verify the number of delivery groups
					resource.TestCheckResourceAttr("citrix_application.testApplication", "delivery_groups.#", "1"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_executable", "test.exe"),

					// /*** Verify Policy Set ***/
					// // Verify name of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "name", os.Getenv("TEST_POLICY_SET_NAME")+"-1"),
					// // Verify description of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "description", "Test policy set description"),
					// // Verify type of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "type", "DeliveryGroupPolicies"),
					// // Verify the number of scopes of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "scopes.#", "0"),
					// // Verify the number of policies in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.#", "2"),
					// // Verify name of the first policy in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.name", "first-test-policy"),
					// // Verify policy settings of the first policy in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.#", "2"),
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.0.name", "AdvanceWarningPeriod"),
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.0.value", "13:00:00"),
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.1.name", "AllowFileDownload"),
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.policy_settings.1.enabled", "true"),
					// // Verify name of the second policy in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.1.name", "second-test-policy"),

					/*** Verify Admin Scope ***/
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "name", adminScopeName),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "description", "test scope created via terraform"),

					/*** Verify Admin Role ***/
					// Verify the name of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "name", adminRoleName),
					// Verify the description of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "description", "Test role created via terraform"),
					// Verify the value of the can_launch_manage flag (Set to true by default)
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "can_launch_manage", "true"),
					// Verify the value of the can_launch_monitor flag (Set to true by default)
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "can_launch_monitor", "true"),
					// Verify the permissions list
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_role.test_role", "permissions.*", "Director_DismissAlerts"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_role.test_role", "permissions.*", "DesktopGroup_AddApplicationGroup"),
				),
			},

			// ImportState testing - Application
			{
				ResourceName:      "citrix_application.testApplication",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"delivery_groups", "installed_app_properties"},
			},
			// // ImportState testing - Policy Set
			// {
			// 	ResourceName:      "citrix_policy_set.testPolicySet",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// The last_updated attribute does not exist in the Orchestration
			// 	// API, therefore there is no value for it during import.
			// 	ImportStateVerifyIgnore: []string{"last_updated"},
			// },
			// ImportState testing - Admin Scope
			{
				ResourceName:      "citrix_admin_scope.test_scope",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// ImportState testing - Admin Role
			{
				ResourceName:      "citrix_admin_role.test_role",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "permissions"},
			},

			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminRoleResource(t, adminRoleTestResource_updated),
					BuildAdminScopeResource(t, adminScopeTestResource_updated),
					// BuildPolicySetResource(t, policy_set_updated_testResource),
					BuildApplicationResource(t, testApplicationResource_updated),
					BuildAdminFolderResourceWithTwoTypes(t, testAdminFolderResource_twoTypes, "ContainsMachineCatalogs", "ContainsApplications"),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					/*** Verify Application ***/
					// Verify name of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "name", fmt.Sprintf("%s-updated", applicationName)),
					// Verify description of application
					resource.TestCheckResourceAttr("citrix_application.testApplication", "description", "Application for testing updated"),
					// Verify the command line arguments
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_arguments", "update test arguments"),
					// Verify the command line executable
					resource.TestCheckResourceAttr("citrix_application.testApplication", "installed_app_properties.command_line_executable", "updated_test.exe"),
					// Verify the application folder path
					resource.TestCheckResourceAttr("citrix_application.testApplication", "application_folder_path", fmt.Sprintf("%s\\", folder_name_2)),

					// /*** Verify Policy Set ***/
					// // Verify name of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "name", os.Getenv("TEST_POLICY_SET_NAME")+"-3"),
					// // Verify description of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "description", "Test policy set description updated"),
					// // Verify type of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "type", "DeliveryGroupPolicies"),
					// // Verify the number of scopes of the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "scopes.#", "0"),
					// // Verify the number of policies in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.#", "1"),
					// // Verify name of the second policy in the policy set
					// resource.TestCheckResourceAttr("citrix_policy_set.testPolicySet", "policies.0.name", "first-test-policy"),

					/*** Verify Admin Scope ***/
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "name", fmt.Sprintf("%s-updated", adminScopeName)),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "description", "Updated description for test scope"),

					/*** Verify Admin Role ***/
					// Verify the name of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "name", fmt.Sprintf("%s-updated", adminRoleName)),
					// Verify the description of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "description", "Updated description for test role"),
					// Verify the value of the can_launch_manage flag
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "can_launch_manage", "true"),
					// Verify the value of the can_launch_monitor flag
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "can_launch_monitor", "true"),
					// Verify the permissions list
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_role.test_role", "permissions.*", "Director_DismissAlerts"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_role.test_role", "permissions.*", "ApplicationGroup_AddScope"),
					resource.TestCheckTypeSetElemAttr("citrix_admin_role.test_role", "permissions.*", "AppLib_AddPackage"),
				),
			},

			/****************** Admin User Test - On-Premises test only ******************/
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminUserResource(t, adminUserTestResource),
					BuildAdminRoleResource(t, adminRoleTestResource_updated),
					BuildAdminScopeResource(t, adminScopeTestResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "name", userName),
					// Verify the domain of the admin use
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "domain_name", userDomainName),
					// Verify the rights object
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.#", "1"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.role", fmt.Sprintf("%s-updated", adminRoleName)),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.scope", fmt.Sprintf("%s-updated", adminScopeName)),
					// Verify the is_enabled flag
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "is_enabled", "true"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_admin_user.test_admin_user",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
				SkipFunc:                skipForCloud(isOnPremises),
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminUserResource(t, adminUserTestResource_updated),
					BuildAdminRoleResource(t, adminRoleTestResource_updated),
					BuildAdminScopeResource(t, adminScopeTestResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_delete_machine, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
					BuildZoneResource(t, zoneInput, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "name", userName),
					// Verify the domain of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "domain_name", userDomainName),
					// Verify the updated rights object
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.#", "2"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.role", fmt.Sprintf("%s-updated", adminRoleName)),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.scope", fmt.Sprintf("%s-updated", adminScopeName)),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.1.role", "Delivery Group Administrator"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.1.scope", fmt.Sprintf("%s-updated", adminScopeName)),
					// Verify the is_enabled flag
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "is_enabled", "true"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
		},
	})
}

func composeAzureMachineCatalogTestResourceTf(t *testing.T, isOnPremises bool) string {
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")

	if isOnPremises {
		return composeTestResourceTf(
			BuildMachineCatalogResourceManualPowerManagedAzure(t, machinecatalog_testResources_manual_power_managed_azure),
			BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "-HybAAD", "HybridAzureAD"),
			BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "", "ActiveDirectory"),
			BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
			BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
			BuildZoneResource(t, zoneInput, true),
		)
	}
	return composeTestResourceTf(
		BuildMachineCatalogResourceManualPowerManagedAzure(t, machinecatalog_testResources_manual_power_managed_azure),
		BuildMachineCatalogResourceWorkgroup(t, machinecatalog_testResources_workgroup),
		BuildMachineCatalogResourceAzureAd(t, machinecatalog_testResources_azure_ad),
		BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "-HybAAD", "HybridAzureAD"),
		BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure, "", "ActiveDirectory"),
		BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_updated_testResource_azure),
		BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
		BuildZoneResource(t, zoneInput, true),
	)
}

func composeCloudMachineCatalogTestVerification(isOnPremises bool, aadCatalogName, workgroupCatalogName string) resource.TestCheckFunc {
	if isOnPremises {
		// For OnPremises, do not return any cloud related test verification
		return resource.ComposeAggregateTestCheckFunc()
	}
	return resource.ComposeAggregateTestCheckFunc(
		/*** Verify MCS AAD Machine Catalog ***/
		// Verify name of catalog
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "name", aadCatalogName),
		// Verify domain FQDN
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "session_support", "MultiSession"),
		// Verify machine catalog identity type
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.identity_type", "AzureAD"),
		// Verify nic network
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-AAD", "provisioning_scheme.network_mapping.0.network", os.Getenv("TEST_MC_SUBNET")),

		/*** Verify MCS Workgroup Machine Catalog ***/
		// Verify name of catalog
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "name", workgroupCatalogName),
		// Verify domain FQDN
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "session_support", "MultiSession"),
		// Verify machine catalog identity type
		resource.TestCheckResourceAttr("citrix_machine_catalog.testMachineCatalog-WG", "provisioning_scheme.identity_type", "Workgroup"),
	)
}
