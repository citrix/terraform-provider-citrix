// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_admin_user", &resource.Sweeper{
		Name: "citrix_admin_user",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)
			if client.ClientConfig.CustomerId != "CitrixOnPremises" {
				// Admin user is not supported for cloud customers
				return nil
			}

			adminUserName := os.Getenv("TEST_ADMIN_USER_NAME")
			err := adminUserSweeper(ctx, adminUserName, client)
			return err
		},
	})
}

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
	zoneInput := os.Getenv("TEST_ZONE_INPUT_AZURE")
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
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsOnly"),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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
					BuildDeliveryGroupResource(t, testDeliveryGroupResources, "DesktopsOnly"),
					BuildMachineCatalogResourceAzure(t, machinecatalog_testResources_azure_updated, "", "ActiveDirectory"),
					BuildHypervisorResourcePoolResourceAzure(t, hypervisor_resource_pool_testResource_azure),
					BuildHypervisorResourceAzure(t, hypervisor_testResources),
					BuildZoneResource(t, zoneInput, false),
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

func adminUserSweeper(ctx context.Context, adminUserName string, client *citrixclient.CitrixDaasClient) error {
	getAdminUserRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminAdministrator(ctx, adminUserName)
	adminUser, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.AdministratorResponseModel](getAdminUserRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting admin user: %s", err)
	}
	userDetails := adminUser.GetUser()
	deleteAdminUserRequest := client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminAdministrator(ctx, userDetails.GetSid())
	httpResp, err = citrixclient.AddRequestData(deleteAdminUserRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", adminUserName, err)
	}
	return nil
}
