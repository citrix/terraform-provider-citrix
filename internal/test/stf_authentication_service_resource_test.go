package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSTFAuthenticationServicePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_STF_SITE_ID"); v == "" {
		t.Fatal("TEST_STF_SITE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_STF_Auth_Virtual_Path"); v == "" {
		t.Fatal("TEST_STF_Auth_Virtual_Path must be set for acceptance tests")
	}
}

func TestSTFAuthenticationServiceResource(t *testing.T) {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	virtualPath := os.Getenv("TEST_STF_Auth_Virtual_Path")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestSTFDeploymentPreCheck(t)
			TestSTFAuthenticationServicePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildSTFAuthenticationServiceResource(t, testSTFAuthenticationServiceResources),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the STF authentication service
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "virtual_path", virtualPath),
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "friendly_name", "testSTFAuthenticationService"),
				),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_stf_authentication_service.testSTFAuthenticationService",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "virtual_path",
				ImportStateIdFunc:                    generateImportStateId_STFAuthenService,
				ImportStateVerifyIgnore:              []string{"last_updated"},
			},

			// Update testing for STF authentication service
			{
				Config: BuildSTFAuthenticationServiceResource(t, testSTFAuthenticationServiceResources_updated),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify parameters of the updated STF authentication service
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "site_id", siteId),
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "virtual_path", virtualPath),
					resource.TestCheckResourceAttr("citrix_stf_authentication_service.testSTFAuthenticationService", "friendly_name", "testAuthServiceUpdated")),
			},
		},
	})
}

func BuildSTFAuthenticationServiceResource(t *testing.T, authService string) string {
	siteId := os.Getenv("TEST_STF_SITE_ID")
	authVirtualPath := os.Getenv("TEST_STF_Auth_Virtual_Path")
	return BuildSTFDeploymentResource(t, testSTFDeploymentResources, siteId) + fmt.Sprintf(authService, authVirtualPath)
}

func generateImportStateId_STFAuthenService(state *terraform.State) (string, error) {
	resourceName := "citrix_stf_authentication_service.testSTFAuthenticationService"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["site_id"], rawState["virtual_path"]), nil
}

var (
	testSTFAuthenticationServiceResources = `
	resource "citrix_stf_authentication_service" "testSTFAuthenticationService" {
		site_id       = citrix_stf_deployment.testSTFDeployment.site_id
		virtual_path  = "%s"
		friendly_name = "testSTFAuthenticationService"
	}
	`

	testSTFAuthenticationServiceResources_updated = `
	resource "citrix_stf_authentication_service" "testSTFAuthenticationService" {
		site_id       = citrix_stf_deployment.testSTFDeployment.site_id
		virtual_path  = "%s"
		friendly_name = "testAuthServiceUpdated"
	}
	`
)
