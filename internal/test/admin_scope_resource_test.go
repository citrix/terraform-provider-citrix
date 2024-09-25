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
	resource.AddTestSweepers("citrix_admin_scope", &resource.Sweeper{
		Name: "citrix_admin_scope",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			adminScopeName := os.Getenv("TEST_ADMIN_SCOPE_NAME")
			err := adminScopeSweeper(ctx, adminScopeName, client)
			return err
		},
		Dependencies: []string{"citrix_admin_user"},
	})
}

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
			TestAdminScopeResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildAdminScopeResource(t, adminScopeTestResource),
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
				Config: BuildAdminScopeResource(t, adminScopeTestResource_updated),
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

func adminScopeSweeper(ctx context.Context, adminScopeName string, client *citrixclient.CitrixDaasClient) error {
	getAdminScopeRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScope(ctx, adminScopeName)
	adminScope, httpResp, err := citrixclient.ExecuteWithRetry[*citrixorchestration.ScopeResponseModel](getAdminScopeRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting admin scope: %s", err)
	}
	deleteAdminScopeRequest := client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminScope(ctx, adminScope.GetId())
	httpResp, err = citrixclient.AddRequestData(deleteAdminScopeRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", adminScopeName, err)
	}
	return nil
}
