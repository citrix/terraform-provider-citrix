package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSTFStoreFarmPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_FARM_PORT"); v == "" {
		t.Fatal("TEST_STF_FARM_PORT must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_FARM_TYPE"); v == "" {
		t.Fatal("TEST_STF_FARM_TYPE must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_FARM_NAME"); v == "" {
		t.Fatal("TEST_STF_FARM_NAME must be set for acceptance tests")
	}
}

func TestSTFStoreFarmResource(t *testing.T) {
	farmType := os.Getenv("TEST_STF_FARM_TYPE")
	farmPort := os.Getenv("TEST_STF_FARM_PORT")
	farmName := os.Getenv("TEST_STF_FARM_NAME")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFStoreFarmPreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFStoreFarmResource(t, testSTFStoreFarmResources),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the STF Store Farm
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "zones.#", "2"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "servers.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "port", farmPort),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "farm_name", farmName),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "farm_type", farmType),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_store_farm.testSTFStoreFarm",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "farm_name",
				ImportStateIdFunc:                    generateImportStateId_STFStoreFarm,
				ImportStateVerifyIgnore:              []string{"last_updated"},
			},

			// Update testing for STF authentication service
			{
				Config: BuildSTFStoreFarmResource(t, testSTFStoreFarmResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the STF Store Farm
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "zones.#", "3"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "zones.0", "Primary"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "zones.1", "Secondary"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "zones.2", "Thirds"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "servers.#", "1"),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "port", farmPort),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "farm_name", farmName),
					resource.TestCheckResourceAttr("citrix_stf_store_farm.testSTFStoreFarm", "farm_type", farmType),
				),
			},
		},
	})
}

func BuildSTFStoreFarmResource(t *testing.T, storeFarm string) string {
	farmType := os.Getenv("TEST_STF_FARM_TYPE")
	farmPort := os.Getenv("TEST_STF_FARM_PORT")
	farmName := os.Getenv("TEST_STF_FARM_NAME")
	return BuildSTFStoreServiceResource(t, testSTFStoreServiceResources) + fmt.Sprintf(storeFarm, farmName, farmType, farmPort)
}

func generateImportStateId_STFStoreFarm(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_store_farm.testSTFStoreFarm"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["store_virtual_path"], rawState["farm_name"]), nil
}

var (
	testSTFStoreFarmResources = `
	resource "citrix_stf_store_farm" "testSTFStoreFarm" {
		store_virtual_path      = citrix_stf_store_service.testSTFStoreService.virtual_path
		farm_name = "%s"
		farm_type = "%s"
		servers = ["cvad.storefront.com"] 
		port = "%s"
		zones = ["Primary","Secondary"]
	}
	`

	testSTFStoreFarmResources_updated = `
	resource "citrix_stf_store_farm" "testSTFStoreFarm" {
		store_virtual_path      = citrix_stf_store_service.testSTFStoreService.virtual_path
		farm_name = "%s"
		farm_type = "%s"
		servers = ["cvad.storefront.com"] 
		port = "%s"
		zones = ["Primary","Secondary", "Thirds"]
	}
	`
)
