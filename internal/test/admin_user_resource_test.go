// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAdminUserPreCheck(t *testing.T) {
	if name := os.Getenv("TEST_ADMIN_USER_NAME"); name == "" {
		t.Fatal("TEST_ADMIN_USER_NAME must be set for acceptance tests")
	}
	if domainName := os.Getenv("TEST_ADMIN_USER_DOMAIN"); domainName == "" {
		t.Fatal("TEST_ADMIN_USER_DOMAIN must be set for acceptance tests")
	}
}

func TestAdminUserResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}
	userName := os.Getenv("TEST_ADMIN_USER_NAME")
	userDomainName := os.Getenv("TEST_ADMIN_USER_DOMAIN")
	roleName := os.Getenv("TEST_ROLE_NAME")
	scopeName := os.Getenv("TEST_ADMIN_SCOPE_NAME")

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
			TestAdminRolePreCheck(t)
			TestAdminUserPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminUserResource(t, adminUserTestResource),
					BuildAdminScopeResource(t, adminScopeTestResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "name", userName),
					// Verify the domain of the admin use
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "domain_name", userDomainName),
					// Verify the rights object
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.#", "1"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.role", roleName),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.scope", scopeName),
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
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					BuildAdminUserResource(t, adminUserTestResource_updated),
					BuildAdminScopeResource(t, adminScopeTestResource),
					BuildDeliveryGroupResource(t, testDeliveryGroupResources),
					BuildPolicySetResourceWithoutDeliveryGroup(t),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "name", userName),
					// Verify the domain of the admin user
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "domain_name", userDomainName),
					// Verify the updated rights object
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.#", "2"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.role", roleName),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.0.scope", scopeName),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.1.role", "Delivery Group Administrator"),
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "rights.1.scope", scopeName),
					// Verify the is_enabled flag
					resource.TestCheckResourceAttr("citrix_admin_user.test_admin_user", "is_enabled", "true"),
				),
				SkipFunc: skipForCloud(isOnPremises),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	adminUserTestResource = `
	resource "citrix_admin_user" "test_admin_user" {
		name = "%s"
		domain_name = "%s"
		rights = [
			{
				role = citrix_admin_role.test_role.name,
				scope = citrix_admin_scope.test_scope.name
			}
		]
		is_enabled = true
	}
	`
	adminUserTestResource_updated = `
	resource "citrix_admin_user" "test_admin_user" {
		name = "%s"
		domain_name = "%s"
		rights = [
			{
				role = citrix_admin_role.test_role.name,
				scope = citrix_admin_scope.test_scope.name
			},
			{
				role = "Delivery Group Administrator",
				scope = citrix_admin_scope.test_scope.name
			}
		]
		is_enabled = true
	}
	`
)

func BuildAdminUserResource(t *testing.T, adminUser string) string {
	adminName := os.Getenv("TEST_ADMIN_USER_NAME")
	adminDomain := os.Getenv("TEST_ADMIN_USER_DOMAIN")
	return fmt.Sprintf(adminUser, adminName, adminDomain)
}
