// Copyright © 2024. Citrix Systems, Inc.

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
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
			TestHypervisorResourcePoolPreCheck_Azure(t)
			TestMachineCatalogPreCheck_Azure(t)
			TestDeliveryGroupPreCheck(t)
			TestAdminScopeResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminScopeResource(t, adminScopeTestResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "name", name),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "description", "test scope created via terraform"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_admin_scope.test_scope",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminScopeResource(t, adminScopeTestResource_updated),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "name", fmt.Sprintf("%s-updated", name)),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("citrix_admin_scope.test_scope", "description", "Updated description for test scope"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	adminScopeTestResource = `
	resource "citrix_admin_scope" "test_scope" {
		name = "%s"
		description = "test scope created via terraform"
	}
	`
	adminScopeTestResource_updated = `
	resource "citrix_admin_scope" "test_scope" {
		name        = "%s-updated"
		description = "Updated description for test scope"
	}
	`
)

func BuildAdminScopeResource(t *testing.T, adminScope string) string {
	return fmt.Sprintf(adminScope, os.Getenv("TEST_ADMIN_SCOPE_NAME"))
}
