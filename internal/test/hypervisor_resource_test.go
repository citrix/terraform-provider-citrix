// Copyright © 2024. Citrix Systems, Inc.

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
				Config: composeTestResourceTf(BuildHypervisorResourceAzure(t, hypervisor_testResources), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE"))),

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
				Config: composeTestResourceTf(BuildHypervisorResourceAzure(t, hypervisor_testResources_updated), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AZURE"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceGCP(t, hypervisor_testResources_gcp), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_GCP"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceGCP(t, hypervisor_testResources_updated_gcp), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_GCP"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceVsphere(t, hypervisor_testResources_vsphere), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_VSPHERE"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceVsphere(t, hypervisor_testResources_updated_vsphere), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_VSPHERE"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceXenserver(t, hypervisor_testResources_xenserver), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_XENSERVER"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceXenserver(t, hypervisor_testResources_updated_xenserver), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_XENSERVER"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceNutanix(t, hypervisor_testResources_nutanix), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_NUTANIX"))),
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
				Config: composeTestResourceTf(BuildHypervisorResourceNutanix(t, hypervisor_testResources_updated_nutanix), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_NUTANIX"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_nutanix_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_SCVMM(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_SCVMM"); v == "" {
		t.Fatal("TEST_ZONE_NAME_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_NAME_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_USERNAME_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_USERNAME_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_PASSWORD_PLAINTEXT_SCVMM must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_ADDRESS_SCVMM"); v == "" {
		t.Fatal("TEST_HYPERV_ADDRESS_SCVMM must be set for acceptance tests")
	}
}

func TestHypervisorResourceSCVMM(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_SCVMM")
	username := os.Getenv("TEST_HYPERV_USERNAME_SCVMM")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_SCVMM(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(BuildHypervisorResourceSCVMM(t, hypervisor_testResources_scvmm), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_SCVMM"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "name", name),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "username", username),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "addresses.#", "1"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_absolute_active_actions", "50"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_absolute_new_actions_per_minute", "10"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_power_actions_percentage_of_machines", "10"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_scvmm_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"password", "password_format"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(BuildHypervisorResourceSCVMM(t, hypervisor_testResources_updated_scvmm), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_SCVMM"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_absolute_active_actions", "40"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_absolute_new_actions_per_minute", "30"),
					resource.TestCheckResourceAttr("citrix_scvmm_hypervisor.testHypervisor", "max_power_actions_percentage_of_machines", "20"),
				),
			},
		},
	})
}

func TestHypervisorPreCheck_AWS_EC2(t *testing.T) {
	if v := os.Getenv("TEST_ZONE_NAME_AWS_EC2"); v == "" {
		t.Fatal("TEST_ZONE_NAME_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_NAME_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_NAME_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_API_KEY_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_API_KEY_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_SECRET_KEY_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_SECRET_KEY_AWS_EC2 must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_HYPERV_REGION_AWS_EC2"); v == "" {
		t.Fatal("TEST_HYPERV_REGION_AWS_EC2 must be set for acceptance tests")
	}
}

func TestHypervisorResourceAwsEc2(t *testing.T) {
	name := os.Getenv("TEST_HYPERV_NAME_AWS_EC2")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestHypervisorPreCheck_AWS_EC2(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_aws_ec2), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AWS_EC2"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_aws_hypervisor.testHypervisor", "name", name),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_aws_hypervisor.testHypervisor",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"api_key", "secret_key"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(BuildHypervisorResourceAwsEc2(t, hypervisor_testResources_updated_aws_ec2), BuildZoneResource(t, zone_testResource, os.Getenv("TEST_ZONE_NAME_AWS_EC2"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_aws_hypervisor.testHypervisor", "name", fmt.Sprintf("%s-updated", name)),
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

// test resources for SCVMM hypervisor
var (
	hypervisor_testResources_scvmm = `
	resource citrix_scvmm_hypervisor "testHypervisor" {
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

	hypervisor_testResources_updated_scvmm = `
	resource citrix_scvmm_hypervisor "testHypervisor" {
		name                = "%s-updated"
		zone                = %s
		username = "%s"
		password = "%s"
		password_format = "PlainText"
		addresses = [
			"%s"
		]		
		max_absolute_active_actions = 40
		max_absolute_new_actions_per_minute = 30
		max_power_actions_percentage_of_machines = 20
	}
	`
)

// test resources for AWS EC2 hypervisor
var (
	hypervisor_testResources_aws_ec2 = `
	resource citrix_aws_hypervisor "testHypervisor" {
		name                = "%s"
		zone                = %s
    	api_key             = "%s"
    	secret_key          = "%s"
    	region              = "%s"
	}
	`

	hypervisor_testResources_updated_aws_ec2 = `
	resource citrix_aws_hypervisor "testHypervisor" {
		name                = "%s-updated"
		zone                = %s
    	api_key             = "%s"
    	secret_key          = "%s"
    	region              = "%s"
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

	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, tenantId, subscriptionId, applicationSecret, applicationId)
}

func BuildHypervisorResourceGCP(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_GCP")
	serviceAccountId := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_ID")
	serviceAccountCredential := os.Getenv("TEST_HYPERV_SERVICE_ACCOUNT_CREDENTIAL")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, serviceAccountId, serviceAccountCredential)
}

func BuildHypervisorResourceVsphere(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_VSPHERE")
	username := os.Getenv("TEST_HYPERV_USERNAME_VSPHERE")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_VSPHERE")
	address := os.Getenv("TEST_HYPERV_ADDRESS_VSPHERE")
	ssl_thumbprint := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_VSPHERE")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address, ssl_thumbprint)
}

func BuildHypervisorResourceXenserver(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_XENSERVER")
	username := os.Getenv("TEST_HYPERV_USERNAME_XENSERVER")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_XENSERVER")
	address := os.Getenv("TEST_HYPERV_ADDRESS_XENSERVER")
	ssl_thumbprint := os.Getenv("TEST_HYPERV_SSL_THUMBPRINT_XENSERVER")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address, ssl_thumbprint)
}

func BuildHypervisorResourceNutanix(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_NUTANIX")
	username := os.Getenv("TEST_HYPERV_USERNAME_NUTANIX")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_NUTANIX")
	address := os.Getenv("TEST_HYPERV_ADDRESS_NUTANIX")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address)
}

func BuildHypervisorResourceSCVMM(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_SCVMM")
	username := os.Getenv("TEST_HYPERV_USERNAME_SCVMM")
	password := os.Getenv("TEST_HYPERV_PASSWORD_PLAINTEXT_SCVMM")
	address := os.Getenv("TEST_HYPERV_ADDRESS_SCVMM")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, username, password, address)
}

func BuildHypervisorResourceAwsEc2(t *testing.T, hypervisor string) string {
	name := os.Getenv("TEST_HYPERV_NAME_AWS_EC2")
	api_key := os.Getenv("TEST_HYPERV_API_KEY_AWS_EC2")
	secret_key := os.Getenv("TEST_HYPERV_SECRET_KEY_AWS_EC2")
	region := os.Getenv("TEST_HYPERV_REGION_AWS_EC2")
	zoneValueForHypervisor := "citrix_zone.test.id"
	return fmt.Sprintf(hypervisor, name, zoneValueForHypervisor, api_key, secret_key, region)
}
