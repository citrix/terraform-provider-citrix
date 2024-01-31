// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAdminScopeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { TestProviderPreCheck(t) },
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: admin_scope_test_data_source_using_name,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin scope
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "id", "00000000-0000-0000-0000-000000000000"),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "description", "All objects"),
					// Verify the is_built_in attribute of the admin scope (Value should be true for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "is_built_in", "true"),
					// Verify the is_all_scope attribute of the admin scope (Value should be true for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "is_all_scope", "true"),
					// Verify the is_tenant_scope attribute of the admin scope (Value should be false for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "is_tenant_scope", "false"),
					// Verify the tenant_id attribute of the admin scope (Value should be empty for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "tenant_id", ""),
					// Verify the tenant_name attribute of the admin scope (Value should be empty for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_name", "tenant_name", ""),
				),
			},
			// Read testing using Name
			{
				Config: admin_scope_test_data_source_using_id,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the admin scope
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "name", "All"),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "description", "All objects"),
					// Verify the is_built_in attribute of the admin scope (Value should be true for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "is_built_in", "true"),
					// Verify the is_all_scope attribute of the admin scope (Value should be true for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "is_all_scope", "true"),
					// Verify the is_tenant_scope attribute of the admin scope (Value should be false for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "is_tenant_scope", "false"),
					// Verify the tenant_id attribute of the admin scope (Value should be empty for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "tenant_id", ""),
					// Verify the tenant_name attribute of the admin scope (Value should be empty for "All" scope)
					resource.TestCheckResourceAttr("data.citrix_daas_admin_scope.test_scope_by_id", "tenant_name", ""),
				),
			},
		},
	})
}

var (
	admin_scope_test_data_source_using_name = `
	data "citrix_daas_admin_scope" "test_scope_by_name" {
		name = "All"
	}
	`
)

var (
	admin_scope_test_data_source_using_id = `
	data "citrix_daas_admin_scope" "test_scope_by_id" {
		id = "00000000-0000-0000-0000-000000000000"
	}
	`
)
