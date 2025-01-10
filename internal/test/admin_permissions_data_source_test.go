// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAdminPermissionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: BuildAdminPermissionsDataSource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.citrix_admin_permissions.test_all_permissions", "permissions.#", func(val string) error {
						if val == "0" {
							return fmt.Errorf("expected at least one permission")
						}
						return nil
					}),
				),
			},
		},
	})
}

func BuildAdminPermissionsDataSource(t *testing.T) string {
	return adminPermissionsTestDataSource
}

var (
	adminPermissionsTestDataSource = `
	data "citrix_admin_permissions" "test_all_permissions" { }
	`
)
