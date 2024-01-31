// Copyright © 2023. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testHypervisorPreCheck validates the necessary env variable exist
// in the testing environment
func TestHypervisorPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_NAME"); v == "" {
		t.Fatal("TEST_HYPERV_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_AD_ID"); v == "" {
		t.Fatal("TEST_HYPERV_AD_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SUBSCRIPTION_ID"); v == "" {
		t.Fatal("TEST_HYPERV_SUBSCRIPTION_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_APPLICATION_ID"); v == "" {
		t.Fatal("TEST_HYPERV_APPLICATION_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_APPLICATION_SECRET"); v == "" {
		t.Fatal("TEST_HYPERV_APPLICATION_SECRET must be set for acceptance tests")
	}
}

func TestHypervisorResourceAzureRM(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildHypervisorResourceAzure(t, hypervisor_testResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_azure_hypervisor.testHypervisor", "name", name),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_daas_azure_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "application_secret"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResourceAzure(t, hypervisor_testResources_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_azure_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_GCP(t *testing.T) {
	if v := os.Getenv("TEST_HYPERV_NAME_GCP"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_GCP must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_ID"); v == "" {
		t.Fatal("TEST_HYPERV_SERVICE_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_CREDENTIAL"); v == "" {
		t.Fatal("TEST_HYPERV_SERVICE_ACCOUNT_CREDENTIAL must be set for acceptance tests")
	}
}

func TestHypervisorResourceGCP(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_GCP")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestZonePreCheck(t)
			TestHypervisorPreCheck_GCP(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_gcp_hypervisor.testHypervisor", "name", name),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_daas_gcp_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "application_secret", "service_account_credentials"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResourceGCP(t, hypervisor_testResources_updated_gcp),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_daas_gcp_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

// test resources for AzureRM hypervisor
var (
	hypervisor_testResources = `
resource "citrix_daas_azure_hypervisor" "testHypervisor" {
    name                = "%s"
    zone                = %s
    active_directory_id = "%s"
    subscription_id     = "%s"
    application_secret  = "%s"
    application_id      = "%s"
}
`

	hypervisor_testResources_updated = `
resource "citrix_daas_azure_hypervisor" "testHypervisor" {
    name                = "%s-updated"
    zone                = %s
    active_directory_id = "%s"
    subscription_id     = "%s"
    application_secret  = "%s"
    application_id      = "%s"
}
`
)

// test resources for GCP hypervisor
var (
	hypervisor_testResources_gcp = `
resource "citrix_daas_gcp_hypervisor" "testHypervisor" {
	name                = "%s"
	zone                = %s
	service_account_id  = "%s"
	service_account_credentials = <<-EOT
%sEOT
}
`

	hypervisor_testResources_updated_gcp = `
resource "citrix_daas_gcp_hypervisor" "testHypervisor" {
	name                = "%s-updated"
	zone                = %s
	service_account_id  = "%s"
	service_account_credentials = <<-EOT
%sEOT
}
`
)

func BuildHypervisorResourceAzure(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME")
	tenantId := os.Getenv("TEST_HYPERV_AD_ID")
	subscriptionId := os.Getenv("TEST_HYPERV_SUBSCRIPTION_ID")
	applicationSecret := os.Getenv("TEST_HYPERV_APPLICATION_SECRET")
	applicationId := os.Getenv("TEST_HYPERV_APPLICATION_ID")

	zoneValueForHypervisor := "citrix_daas_zone.test.id"

	return BuildZoneResource(t, zone_testResource) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, tenantId, subscriptionId, applicationSecret, applicationId)
}

func BuildHypervisorResourceGCP(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_GCP")
	serviceAccountId := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_ID")
	serviceAccountCredential := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_CREDENTIAL")
	zoneValueForHypervisor := "citrix_daas_zone.test.id"
	resource := BuildZoneResource(t, zone_testResource) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, serviceAccountId, serviceAccountCredential)
	return resource
}
