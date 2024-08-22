// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBearerTokenDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: BuildBearerTokenDataSource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrix_bearer_token.test_bearer_token", "bearer_token"),
				),
			},
		},
	})
}

func BuildBearerTokenDataSource(t *testing.T) string {
	return bearer_token_test_data_source
}

var (
	bearer_token_test_data_source = `
	data "citrix_bearer_token" "test_bearer_token" { }
	`
)
