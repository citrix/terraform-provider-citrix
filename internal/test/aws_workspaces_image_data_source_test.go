// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAWSWorkspacesImageDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_EXPECTED_SESSION_SUPPORT"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_IMAGE_EXPECTED_SESSION_SUPPORT must be set for acceptance tests")
	}
}

func TestAwsWorkspacesImageDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	imageId := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_ID")
	imageName := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_DATA_SOURCE_NAME")
	accountId := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ID")
	sessionSupport := os.Getenv("TEST_AWS_WORKSPACES_IMAGE_EXPECTED_SESSION_SUPPORT")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesImageDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Image ID and Account Id
			{
				Config: BuildAWSWorkspacesImageDataSource(t, aws_workspaces_image_test_data_source_with_id, imageId, accountId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "id", imageId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "name", imageName),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "account_id", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "session_support", sessionSupport),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Read testing using Image Name and Account Id
			{
				Config: BuildAWSWorkspacesImageDataSource(t, aws_workspaces_image_test_data_source_with_name, imageName, accountId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "id", imageId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "name", imageName),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "account_id", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image", "session_support", sessionSupport),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesImageDataSource(t *testing.T, workspacesImageResource string, id string, accountId string) string {
	return fmt.Sprintf(workspacesImageResource, id, accountId)
}

var (
	aws_workspaces_image_test_data_source_with_id = `
	data "citrix_quickcreate_aws_workspaces_image" "test_aws_workspaces_image" {
		id         = "%s"
		account_id = "%s"
	}
	`

	aws_workspaces_image_test_data_source_with_name = `
	data "citrix_quickcreate_aws_workspaces_image" "test_aws_workspaces_image" {
		name       = "%s"
		account_id = "%s"
	}
	`
)
