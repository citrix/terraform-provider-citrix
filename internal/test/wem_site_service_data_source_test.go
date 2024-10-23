// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWemSiteDataSourcePreCheck(t *testing.T) {

	if v := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_WEM_SITE_DATA_SOURCE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_WEM_SITE_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_DESCRIPTION"); v == "" {
		t.Fatal("TEST_WEM_SITE_DATA_SOURCE_DESCRIPTION must be set for acceptance tests")
	}
}

func TestWemSiteDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_ID")
	name := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_NAME")
	description := os.Getenv("TEST_WEM_SITE_DATA_SOURCE_DESCRIPTION")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestWemSiteDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildWemSiteDataSource(t, wem_site_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "id", id),
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "name", name),
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "description", description),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildWemSiteDataSource(t, wem_site_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "id", id),
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "name", name),
					resource.TestCheckResourceAttr("data.citrix_wem_configuration_set.test_wem_site", "description", description),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildWemSiteDataSource(t *testing.T, wemSiteDataSource string, idOrName string) string {
	return fmt.Sprintf(wemSiteDataSource, idOrName)
}

var (
	wem_site_data_source_using_id = `
	data "citrix_wem_configuration_set" "test_wem_site" {
		id         = "%s"
	}
	`

	wem_site_data_source_using_name = `
	data "citrix_wem_configuration_set" "test_wem_site" {
		name       = "%s"
	}
	`
)
