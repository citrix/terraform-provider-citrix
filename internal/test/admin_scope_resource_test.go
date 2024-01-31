// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAdminScopeResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestAdminScopeResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_ADMIN_SCOPE_NAME"); v == "" {
		t.Fatal("TEST_ADMIN_SCOPE_NAME must be set for acceptance tests")
	}
}

func TestAdminScopeResource(t *testing.T) {
	name := os.Getenv("TEST_ADMIN_SCOPE_NAME")
	catalogName := os.Getenv("TEST_MC_NAME")
	dgName := os.Getenv("TEST_DG_NAME")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestAdminScopeResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildAdminScopeResource(t, adminScopeTestResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "name", name),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "description", "test scope created via terraform"),
					// Verify number of scoped objects
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.#", "1"),
					// Verify the scoped objects data
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.0.object_type", "DeliveryGroup"),
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.0.object", dgName),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_admin_scope.test_scope",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: BuildAdminScopeResource(t, adminScopeTestResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "name", fmt.Sprintf("%s-updated", name)),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "description", "Updated description for test scope"),
					// Verify number of scoped objects
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.#", "2"),
					// Verify the scoped objects data
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.0.object_type", "DeliveryGroup"),
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.0.object", dgName),
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.1.object_type", "MachineCatalog"),
					resource.TestCheckResourceAttr("citrix_daas_admin_scope.test_scope", "scoped_objects.1.object", catalogName),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	adminScopeTestResource = `
	resource "citrix_daas_admin_scope" "test_scope" {
		name = "%s"
		description = "test scope created via terraform"
		scoped_objects = [
			{
				object_type = "DeliveryGroup",
				object = citrix_daas_delivery_group.testDeliveryGroup.name
			}
		]
	}
	`
	adminScopeTestResource_updated = `
	resource "citrix_daas_admin_scope" "test_scope" {
		name        = "%s-updated"
		description = "Updated description for test scope"
		scoped_objects    = [
			{
				object_type = "DeliveryGroup",
				object = citrix_daas_delivery_group.testDeliveryGroup.name
			},
			{
				object_type = "MachineCatalog",
				object = citrix_daas_machine_catalog.testMachineCatalog.name
			}
		]
	}
	`
)

func BuildAdminScopeResource(t *testing.T, adminScope string) string {
	return BuildDeliveryGroupResource(t, testDeliveryGroupResources) + fmt.Sprintf(adminScope, os.Getenv("TEST_ADMIN_SCOPE_NAME"))
}
