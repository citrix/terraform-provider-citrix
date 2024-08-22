// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSiteDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_SITE_DATA_SOURCE_EXPECTED_SITE_ID"); v == "" {
		t.Fatal("TEST_SITE_DATA_SOURCE_EXPECTED_SITE_ID must be set for acceptance tests")
	}
}

func TestSiteDataSource(t *testing.T) {
	expectedCustomerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if expectedCustomerId != "" && expectedCustomerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	} else {
		expectedCustomerId = "CitrixOnPremises"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSiteDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: BuildSiteDataSource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_site.test_citrix_site", "site_id", os.Getenv("TEST_SITE_DATA_SOURCE_EXPECTED_SITE_ID")),
					resource.TestCheckResourceAttr("data.citrix_site.test_citrix_site", "customer_id", expectedCustomerId),
					resource.TestCheckResourceAttr("data.citrix_site.test_citrix_site", "is_on_premises", fmt.Sprintf("%t", isOnPremises)),
				),
			},
		},
	})
}

func BuildSiteDataSource(t *testing.T) string {
	return site_test_data_source
}

var (
	site_test_data_source = `
	data "citrix_site" "test_citrix_site" { }
	`
)
