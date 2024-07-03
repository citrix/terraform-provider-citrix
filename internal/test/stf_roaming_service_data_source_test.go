// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSTFRoamingServicePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}
}

func TestSTFRoamingServiceDataSource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFRoamingServicePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId),
			},
			// Read testing using site_id
			{
				Config: BuildSTFRoamingServiceDataSource(t, stf_roaming_service_test_data_source),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the Site ID of the StoreFront Roaming Service
					resource.TestCheckResourceAttr("data.citrix_stf_roaming_service.test_stf_roaming_service", "site_id", siteId),
					resource.TestCheckResourceAttr("data.citrix_stf_roaming_service.test_stf_roaming_service", "name", "Roaming"),
					resource.TestCheckResourceAttr("data.citrix_stf_roaming_service.test_stf_roaming_service", "virtual_path", "/Citrix/Roaming"),
				),
			},
		},
	})
}

func BuildSTFRoamingServiceDataSource(t *testing.T, roamingService string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")

	return BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId) + fmt.Sprintf(roamingService, siteId)
}

var (
	stf_roaming_service_test_data_source = `
	data "citrix_stf_roaming_service" "test_stf_roaming_service" {
		site_id = "%s"
	}
	`
)
