// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAdminRolePreCheck(t *testing.T) {
	if name := os.Getenv("TEST_ROLE_NAME"); name == "" {
		t.Fatal("TEST_ROLE_NAME must be set for acceptance tests")
	}
}

func TestAdminRoleResource(t *testing.T) {
	name := os.Getenv("TEST_ROLE_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAdminRolePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildAdminRoleResource(t, adminRoleTestResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "name", name),
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
			// ImportState testing
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
				Config: BuildAdminRoleResource(t, adminRoleTestResource_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin role
					resource.TestCheckResourceAttr("citrix_admin_role.test_role", "name", fmt.Sprintf("%s-updated", name)),
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
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	adminRoleTestResource = `
	resource "citrix_admin_role" "test_role" {
		name = "%s"
		description = "Test role created via terraform"
		permissions = ["Director_DismissAlerts", "DesktopGroup_AddApplicationGroup"]	
	}
	`
	adminRoleTestResource_updated = `
	resource "citrix_admin_role" "test_role" {
		name = "%s-updated"
		description = "Updated description for test role"
		can_launch_manage = true
		can_launch_monitor = true
		permissions = ["Director_DismissAlerts", "ApplicationGroup_AddScope", "AppLib_AddPackage"]
	}
	`
)

func BuildAdminRoleResource(t *testing.T, adminRole string) string {
	return fmt.Sprintf(adminRole, os.Getenv("TEST_ROLE_NAME"))
}
