// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSTFUserFarmMappingResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_USER_FARM_MAPPING_NAME"); v == "" {
		t.Fatal("TEST_STF_USER_FARM_MAPPING_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER1_SID"); v == "" {
		t.Fatal("TEST_STF_USER_FARM_MAPPING_USER1_SID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER2_SID"); v == "" {
		t.Fatal("TEST_STF_USER_FARM_MAPPING_USER2_SID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_PRIMARY_FARM_NAME"); v == "" {
		t.Fatal("TEST_STF_PRIMARY_FARM_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_SECONDARY_FARM_NAME"); v == "" {
		t.Fatal("TEST_STF_SECONDARY_FARM_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_BACKUP_FARM_NAME"); v == "" {
		t.Fatal("TEST_STF_BACKUP_FARM_NAME must be set for acceptance tests")
	}
}

func TestSTFUserFarmMappingResource(t *testing.T) {
	virtualPath := os.Getenv("TEST_STF_Store_Virtual_Path")
	name := os.Getenv("TEST_STF_USER_FARM_MAPPING_NAME")
	user1Sid := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER1_SID")
	user2Sid := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER2_SID")
	primaryFarmName := os.Getenv("TEST_STF_PRIMARY_FARM_NAME")
	secondaryFarmName := os.Getenv("TEST_STF_SECONDARY_FARM_NAME")
	backupFarmName := os.Getenv("TEST_STF_BACKUP_FARM_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFStoreServicePreCheck(t)
			TestSTFUserFarmMappingResourcePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFUserFarmMappingResource(t, testSTFUserFarmMappingResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "name", name),
					// Verify store_virtual_path of STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "store_virtual_path", virtualPath),
					// Verify group_members for STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.0.group_name", "TestGroup1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.0.account_sid", user1Sid),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.1.group_name", "TestGroup2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.1.account_sid", user2Sid),
					// Verify equivalent_farm_sets of STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.name", "TestEFS1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.aggregation_group_name", "EFS1Group"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.load_balance_mode", "LoadBalanced"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.farms_are_identical", "false"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.primary_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.primary_farms.0", primaryFarmName),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.backup_farms.#", "0"),

					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.name", "TestEFS2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.aggregation_group_name", "EFS2Group"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.load_balance_mode", "Failover"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.farms_are_identical", "true"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.primary_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.primary_farms.0", secondaryFarmName),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.backup_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.backup_farms.0", backupFarmName),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateIdFunc:                    generateImportStateId_STFUserFarmMappingResource,
				ImportStateVerifyIgnore:              []string{"last_updated"},
			},

			// Update testing for STF WebReceiver Service
			{
				Config: BuildSTFUserFarmMappingResource(t, testSTFUserFarmMappingResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "name", fmt.Sprintf("%s-updated", name)),
					// Verify store_virtual_path of STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "store_virtual_path", virtualPath),
					// Verify group_members for STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.0.group_name", "TestGroup2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.0.account_sid", user1Sid),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.1.group_name", "TestGroup1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "group_members.1.account_sid", user2Sid),
					// Verify equivalent_farm_sets of STF UserFarmMapping Resource
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.name", "TestEFS2"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.aggregation_group_name", "EFS2Group"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.load_balance_mode", "Failover"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.farms_are_identical", "true"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.primary_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.primary_farms.0", primaryFarmName),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.0.backup_farms.#", "0"),

					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.name", "TestEFS1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.aggregation_group_name", "EFS1Group"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.load_balance_mode", "LoadBalanced"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.farms_are_identical", "false"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.primary_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.primary_farms.0", secondaryFarmName),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.backup_farms.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource", "equivalent_farm_sets.1.backup_farms.0", backupFarmName),
				),
			},
		},
	})
}

func BuildSTFUserFarmMappingResource(t *testing.T, userFarmMappingResource string) string {
	name := os.Getenv("TEST_STF_USER_FARM_MAPPING_NAME")
	user1Sid := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER1_SID")
	user2Sid := os.Getenv("TEST_STF_USER_FARM_MAPPING_USER2_SID")
	primaryFarmName := os.Getenv("TEST_STF_PRIMARY_FARM_NAME")
	secondaryFarmName := os.Getenv("TEST_STF_SECONDARY_FARM_NAME")
	backupFarmName := os.Getenv("TEST_STF_BACKUP_FARM_NAME")

	return BuildSTFStoreServiceResource(t, testSTFStoreServiceResources) + fmt.Sprintf(userFarmMappingResource, name, user1Sid, user2Sid, primaryFarmName, secondaryFarmName, backupFarmName)
}

func generateImportStateId_STFUserFarmMappingResource(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_user_farm_mapping.testSTFUserFarmMappingResource"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["store_virtual_path"], rawState["name"]), nil
}

var (
	testSTFUserFarmMappingResources = `
	resource "citrix_stf_user_farm_mapping" "testSTFUserFarmMappingResource" {
		name = "%s"
		store_virtual_path = citrix_stf_store_service.testSTFStoreService.virtual_path
		group_members = [
			{
				group_name = "TestGroup1"
				account_sid = "%s"
			},
			{
				group_name = "TestGroup2"
				account_sid = "%s"
			}
		]
		equivalent_farm_sets = [
			{
				name = "TestEFS1",
				aggregation_group_name = "EFS1Group"
				primary_farms = ["%s"]
				load_balance_mode = "LoadBalanced"
				farms_are_identical = false
			},
			{
				name = "TestEFS2",
				aggregation_group_name = "EFS2Group"
				primary_farms = ["%s"]
				backup_farms = ["%s"]
				load_balance_mode = "Failover"
				farms_are_identical = true
			}
		]
	}
	`

	testSTFUserFarmMappingResources_updated = `
	resource "citrix_stf_user_farm_mapping" "testSTFUserFarmMappingResource" {
		name = "%s-updated"
		store_virtual_path = citrix_stf_store_service.testSTFStoreService.virtual_path
		group_members = [
			{
				group_name = "TestGroup2"
				account_sid = "%s"
			},
			{
				group_name = "TestGroup1"
				account_sid = "%s"
			}
		]
		equivalent_farm_sets = [
			{
				name = "TestEFS2",
				aggregation_group_name = "EFS2Group"
				primary_farms = ["%s"]
				load_balance_mode = "Failover"
				farms_are_identical = true
			},
			{
				name = "TestEFS1",
				aggregation_group_name = "EFS1Group"
				primary_farms = ["%s"]
				backup_farms = ["%s"]
				load_balance_mode = "LoadBalanced"
				farms_are_identical = false
			}
		]
	}
	`
)
