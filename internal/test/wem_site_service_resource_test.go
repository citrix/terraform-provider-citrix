// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestWemSiteResourcePreCheck validates the necessary env variable exist in the testing environment
func TestWemSiteResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_WEM_SITE_RESOURCE_NAME"); v == "" {
		t.Fatal("TEST_WEM_SITE_RESOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_WEM_SITE_RESOURCE_DESCRIPTION"); v == "" {
		t.Fatal("TEST_WEM_SITE_RESOURCE_DESCRIPTION must be set for acceptance tests")
	}
}

func TestWemSiteResource(t *testing.T) {
	wemSiteName := os.Getenv("TEST_WEM_SITE_RESOURCE_NAME")
	wemSiteDescription := os.Getenv("TEST_WEM_SITE_RESOURCE_DESCRIPTION")

	wemSiteName_Updated := os.Getenv("TEST_WEM_SITE_RESOURCE_NAME") + "-updated"
	wemSiteDescription_Updated := os.Getenv("TEST_WEM_SITE_RESOURCE_DESCRIPTION") + " description updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestWemSiteResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: BuildWemSiteResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the wem site
					resource.TestCheckResourceAttr("citrix_wem_configuration_set.test_wem_site", "name", wemSiteName),
					// Verify the description of wem site
					resource.TestCheckResourceAttr("citrix_wem_configuration_set.test_wem_site", "description", wemSiteDescription),
				),
			},
			// Import test
			{
				ResourceName:            "citrix_wem_configuration_set.test_wem_site",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update and Read test
			{
				Config: BuildWemSiteResource_Updated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the wem site
					resource.TestCheckResourceAttrSet("citrix_wem_configuration_set.test_wem_site", "id"),
					// Verify the name of the wem site
					resource.TestCheckResourceAttr("citrix_wem_configuration_set.test_wem_site", "name", wemSiteName_Updated),
					// Verify the description of wem site
					resource.TestCheckResourceAttr("citrix_wem_configuration_set.test_wem_site", "description", wemSiteDescription_Updated),
				),
			},
		},
	})
}

func BuildWemSiteResource(t *testing.T) string {
	wemSiteName := os.Getenv("TEST_WEM_SITE_RESOURCE_NAME")
	wemSiteDescription := os.Getenv("TEST_WEM_SITE_RESOURCE_DESCRIPTION")

	return fmt.Sprintf(wem_site_test_resource, wemSiteName, wemSiteDescription)
}

func BuildWemSiteResource_Updated(t *testing.T) string {
	wemSiteName := os.Getenv("TEST_WEM_SITE_RESOURCE_NAME") + "-updated"
	wemSiteDescription := os.Getenv("TEST_WEM_SITE_RESOURCE_DESCRIPTION") + " description updated"

	return fmt.Sprintf(wem_site_test_resource, wemSiteName, wemSiteDescription)
}

var (
	wem_site_test_resource = `
	resource "citrix_wem_configuration_set" "test_wem_site" {
		name = "%s"
		description = "%s"
	}
	`
)
