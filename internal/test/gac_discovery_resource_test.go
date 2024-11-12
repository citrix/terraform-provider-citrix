// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGacDiscoveryPreCheck(t *testing.T) {
	if domain := os.Getenv("TEST_GAC_DISCOVERY_DOMAIN"); domain == "" {
		t.Fatal("TEST_GAC_DISCOVERY_DOMAIN must be set for acceptance tests")
	}
	if service_url := os.Getenv("TEST_GAC_DISCOVERY_SERVICE_URL"); service_url == "" {
		t.Fatal("TEST_GAC_DISCOVERY_SERVICE_URL must be set for acceptance tests")
	}
	if service_url_updated := os.Getenv("TEST_GAC_DISCOVERY_SERVICE_URL_UPDATE"); service_url_updated == "" {
		t.Fatal("TEST_GAC_DISCOVERY_SERVICE_URL_UPDATE must be set for acceptance tests")
	}
}

func TestGacDiscoveryResource(t *testing.T) {
	domain := os.Getenv("TEST_GAC_DISCOVERY_DOMAIN")
	service_url := os.Getenv("TEST_GAC_DISCOVERY_SERVICE_URL")
	service_url_updated := os.Getenv("TEST_GAC_DISCOVERY_SERVICE_URL_UPDATE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestGacDiscoveryPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildGacDiscoveryResource(t, gacDiscoveryTestResource, service_url),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the domain
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "domain", domain),
					// Verify the service urls
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "service_urls.#", "1"),
					// Verify the service urls value
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "service_urls.0", service_url),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "citrix_gac_discovery.test_gac_discovery",
				ImportState:                          true,
				ImportStateId:                        domain,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
			// Update and Read testing
			{
				Config: BuildGacDiscoveryResource(t, gacDiscoveryTestResource_updated, service_url_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the domain after update
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "domain", domain),
					// Verify the service urls
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "service_urls.#", "1"),
					// Verify the service urls value
					resource.TestCheckResourceAttr("citrix_gac_discovery.test_gac_discovery", "service_urls.0", service_url_updated),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	gacDiscoveryTestResource = `
	resource "citrix_gac_discovery" "test_gac_discovery" {
		domain = "%s"
    	service_urls = ["%s"]
	}
	`
	gacDiscoveryTestResource_updated = `
	resource "citrix_gac_discovery" "test_gac_discovery" {
		domain = "%s"
		service_urls = ["%s"]
	}
	`
)

func BuildGacDiscoveryResource(t *testing.T, settings string, serviceURL string) string {
	val := fmt.Sprintf(settings, os.Getenv("TEST_GAC_DISCOVERY_DOMAIN"), serviceURL)
	return val
}
