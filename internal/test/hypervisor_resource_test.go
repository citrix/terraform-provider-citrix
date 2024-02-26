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
	if v := os.Getenv("TEST_ZONE_NAME_AZURE"); v == "" {
		t.Fatal("TEST_ZONE_NAME_AZURE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_NAME_AZURE"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_AZURE must be set for acceptance tests")
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
	name := os.Getenv("TEST_HYPERV_NAME_AZURE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildHypervisorResourceAzure(t, hypervisor_testResources),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_azure_hypervisor.testHypervisor", "name", name),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_azure_hypervisor.testHypervisor",
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
					resource.TestCheckResourceAttr("citrix_azure_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_GCP(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_GCP"); v == "" {
		t.Fatal("TEST_ZONE_NAME_GCP must be set for acceptance tests")
	}
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
			TestHypervisorPreCheck_GCP(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor.testHypervisor", "name", name),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_gcp_hypervisor.testHypervisor",
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
					resource.TestCheckResourceAttr("citrix_gcp_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_Vsphere(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_VSPHERE"); v == "" {
		t.Fatal("TEST_ZONE_NAME_VSPHERE must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_HYPERV_NAME_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_USERNAME_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_USERNAME_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_PASSWORD_PLAINTEXT_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_ADDRESS_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_ADDRESS_VSPHERE must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_VSPHERE"); v == "" {
		t.Fatal("TEST_HYPERV_SSL_THUMBPRINT_VSPHERE must be set for acceptance tests")
	}
}

func TestHypervisorResourceVsphere(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_VSPHERE")
	username := os.Getenv("TEST_HYPERV_USERNAME_VSPHERE")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Vsphere(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor.testHypervisor", "name", name),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor.testHypervisor", "username", username),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor.testHypervisor", "addresses.#", "1"),
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor.testHypervisor", "ssl_thumbprints.#", "1"),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_vsphere_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"password", "password_format"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResourceVsphere(t, hypervisor_testResources_updated_vsphere),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_vsphere_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_Xenserver(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_XENSERVER"); v == "" {
		t.Fatal("TEST_ZONE_NAME_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_NAME_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_USERNAME_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_USERNAME_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_PASSWORD_PLAINTEXT_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_ADDRESS_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_ADDRESS_XENSERVER must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_XENSERVER"); v == "" {
		t.Fatal("TEST_HYPERV_SSL_THUMBPRINT_XENSERVER must be set for acceptance tests")
	}
}

func TestHypervisorResourceXenserver(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_XENSERVER")
	username := os.Getenv("TEST_HYPERV_USERNAME_XENSERVER")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Xenserver(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor.testHypervisor", "name", name),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor.testHypervisor", "username", username),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor.testHypervisor", "addresses.#", "1"),
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor.testHypervisor", "ssl_thumbprints.#", "1"),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_xenserver_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"password", "password_format"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResourceXenserver(t, hypervisor_testResources_updated_xenserver),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_xenserver_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_Nutanix(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_NUTANIX"); v == "" {
		t.Fatal("TEST_ZONE_NAME_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_NAME_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_USERNAME_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_USERNAME_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_PASSWORD_PLAINTEXT_NUTANIX must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_ADDRESS_NUTANIX"); v == "" {
		t.Fatal("TEST_HYPERV_ADDRESS_NUTANIX must be set for acceptance tests")
	}
}

func TestHypervisorResourceNutanix(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_NUTANIX")
	username := os.Getenv("TEST_HYPERV_USERNAME_NUTANIX")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_Nutanix(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor.testHypervisor", "name", name),
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor.testHypervisor", "username", username),
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor.testHypervisor", "addresses.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_nutanix_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"password", "password_format"},
			},
			// Update and Read testing
			{
				Config: BuildHypervisorResourceNutanix(t, hypervisor_testResources_updated_nutanix),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

// test resources for AzureRM hypervisor
var (
	hypervisor_testResources = `
resource "citrix_azure_hypervisor" "testHypervisor" {
    name                = "%s"
    zone                = %s
    active_directory_id = "%s"
    subscription_id     = "%s"
    application_secret  = "%s"
    application_id      = "%s"
}
`

	hypervisor_testResources_updated = `
resource "citrix_azure_hypervisor" "testHypervisor" {
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
resource "citrix_gcp_hypervisor" "testHypervisor" {
	name                = "%s"
	zone                = %s
	service_account_id  = "%s"
	service_account_credentials = <<-EOT
%sEOT
}
`

	hypervisor_testResources_updated_gcp = `
resource "citrix_gcp_hypervisor" "testHypervisor" {
	name                = "%s-updated"
	zone                = %s
	service_account_id  = "%s"
	service_account_credentials = <<-EOT
%sEOT
}
`
)

// test resources for VSPHERE hypervisor
var (
	hypervisor_testResources_vsphere = `
	resource citrix_vsphere_hypervisor "testHypervisor" {
		name                = "%s"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
		ssl_thumbprints = [
			"%s"
		]
	}
	`

	hypervisor_testResources_updated_vsphere = `
	resource citrix_vsphere_hypervisor "testHypervisor" {
		name                = "%s-updated"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
		ssl_thumbprints = [
			"%s"
		]
	}
	`
)

// test resources for XenServer hypervisor
var (
	hypervisor_testResources_xenserver = `
	resource citrix_xenserver_hypervisor "testHypervisor" {
		name                = "%s"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
		ssl_thumbprints = [
			"%s"
		]
	}
	`

	hypervisor_testResources_updated_xenserver = `
	resource citrix_xenserver_hypervisor "testHypervisor" {
		name                = "%s-updated"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
		ssl_thumbprints = [
			"%s"
		]
	}
	`
)

// test resources for Nutanix hypervisor
var (
	hypervisor_testResources_nutanix = `
	resource citrix_nutanix_hypervisor "testHypervisor" {
		name                = "%s"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
	}
	`

	hypervisor_testResources_updated_nutanix = `
	resource citrix_nutanix_hypervisor "testHypervisor" {
		name                = "%s-updated"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]
	}
	`
)

func BuildHypervisorResourceAzure(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_AZURE")
	tenantId := os.Getenv("TEST_HYPERV_AD_ID")
	subscriptionId := os.Getenv("TEST_HYPERV_SUBSCRIPTION_ID")
	applicationSecret := os.Getenv("TEST_HYPERV_APPLICATION_SECRET")
	applicationId := os.Getenv("TEST_HYPERV_APPLICATION_ID")
	zoneValueForHypervisor := "citrix_zone.test.id"

	zoneNameAzure := os.Getenv("TEST_ZONE_NAME_AZURE")
	return BuildZoneResource(t, zone_testResource, zoneNameAzure) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, tenantId, subscriptionId, applicationSecret, applicationId)
}

func BuildHypervisorResourceGCP(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_GCP")
	serviceAccountId := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_ID")
	serviceAccountCredential := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_CREDENTIAL")
	zoneValueForHypervisor := "citrix_zone.test.id"
	zoneNameGCP := os.Getenv("TEST_ZONE_NAME_GCP")
	resource := BuildZoneResource(t, zone_testResource, zoneNameGCP) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, serviceAccountId, serviceAccountCredential)
	return resource
}

func BuildHypervisorResourceVsphere(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_VSPHERE")
	username := os.Getenv("TEST_HYPERV_USERNAME_VSPHERE")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_VSPHERE")
	address := os.Getenv("TEST_HYPERV_ADDRESS_VSPHERE")
	ssl_thumbprint := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_VSPHERE")
	zoneValueForHypervisor := "citrix_zone.test.id"
	zoneNameVsphere := os.Getenv("TEST_ZONE_NAME_VSPHERE")
	return BuildZoneResource(t, zone_testResource, zoneNameVsphere) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address, ssl_thumbprint)
}

func BuildHypervisorResourceXenserver(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_XENSERVER")
	username := os.Getenv("TEST_HYPERV_USERNAME_XENSERVER")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_XENSERVER")
	address := os.Getenv("TEST_HYPERV_ADDRESS_XENSERVER")
	ssl_thumbprint := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_XENSERVER")
	zoneValueForHypervisor := "citrix_zone.test.id"
	zoneNameXenserver := os.Getenv("TEST_ZONE_NAME_XENSERVER")
	return BuildZoneResource(t, zone_testResource, zoneNameXenserver) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address, ssl_thumbprint)
}

func BuildHypervisorResourceNutanix(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_NUTANIX")
	username := os.Getenv("TEST_HYPERV_USERNAME_NUTANIX")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_NUTANIX")
	address := os.Getenv("TEST_HYPERV_ADDRESS_NUTANIX")
	zoneValueForHypervisor := "citrix_zone.test.id"
	zoneNameNutanix := os.Getenv("TEST_ZONE_NAME_NUTANIX")
	return BuildZoneResource(t, zone_testResource, zoneNameNutanix) + fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address)
}
