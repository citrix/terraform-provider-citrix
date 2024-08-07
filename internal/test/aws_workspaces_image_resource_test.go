// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAWSWorkspacesImageResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_AWS_IMAGE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_OS must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_NAME_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_AWS_IMAGE_ID_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_OS_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS_UPDATED must be set for acceptance tests")
	}
}

func TestAWSWorkspacesImageResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	name := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME")
	amazonImageId := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID")
	sessionSupport := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT")
	operatingSystem := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS")
	ingestionProcess := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS")

	name_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME_UPDATED")
	amazonImageId_updated := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID_UPDATED")
	sessionSupport_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT_UPDATED")
	operatingSystem_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS_UPDATED")
	ingestionProcess_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesAccountResourcePreCheck(t)
			TestAWSWorkspacesImageResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Image resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResource(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Workspaces Image Tests
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "name", name),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "aws_image_id", amazonImageId),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "description", "Test Citrix AWS Workspaces Image"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "session_support", sessionSupport),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "operating_system", operatingSystem),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "ingestion_process", ingestionProcess),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},

			// ImportState testing
			{
				ResourceName:                         "citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateIdFunc:                    generateImportStateId_QcsAwsImage,
				ImportStateVerifyIgnore:              []string{},
				SkipFunc:                             skipForOnPrem(isOnPremises),
			},

			// Update testing for QCS AWS Workspaces Image resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResourceUpdated(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "name", name_updated),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "aws_image_id", amazonImageId_updated),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "description", "Test Citrix AWS Workspaces Image Updated"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "session_support", sessionSupport_updated),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "operating_system", operatingSystem_updated),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "ingestion_process", ingestionProcess_updated),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesImageResource(t *testing.T) string {
	name := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME")
	amazonImageId := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID")
	sessionSupport := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT")
	operatingSystem := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS")
	ingestionProcess := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS")
	return fmt.Sprintf(testAWSWorkspacesImageResource, name, amazonImageId, sessionSupport, operatingSystem, ingestionProcess)
}

func BuildAWSWorkspacesImageResourceUpdated(t *testing.T) string {
	name_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_NAME_UPDATED")
	amazonImageId_updated := os.Getenv("TEST_AWS_WORKSPACES_AWS_IMAGE_ID_UPDATED")
	sessionSupport_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_SESSION_SUPPORT_UPDATED")
	operatingSystem_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_OS_UPDATED")
	ingestionProcess_updated := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_INGESTION_PROCESS_UPDATED")

	return fmt.Sprintf(testAWSWorkspacesImageResource_updated, name_updated, amazonImageId_updated, sessionSupport_updated, operatingSystem_updated, ingestionProcess_updated)
}

func generateImportStateId_QcsAwsImage(state *terraform.State) (string, error) {
	resourceName := "citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["account_id"], rawState["id"]), nil
}

var (
	testAWSWorkspacesImageResource = `
	resource "citrix_quickcreate_aws_workspaces_image" "test_aws_workspaces_image" {
		name                       = "%s"
		account_id                 = citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		aws_image_id               = "%s"
		description                = "Test Citrix AWS Workspaces Image"
		session_support            = "%s"
		operating_system           = "%s"
		ingestion_process          = "%s"
	}
	`
	testAWSWorkspacesImageResource_updated = `
	resource "citrix_quickcreate_aws_workspaces_image" "test_aws_workspaces_image" {
		name                       = "%s"
		account_id                 = citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		aws_image_id               = "%s"
		description                = "Test Citrix AWS Workspaces Image Updated"
		session_support            = "%s"
		operating_system           = "%s"
		ingestion_process          = "%s"
	}
	`
)
