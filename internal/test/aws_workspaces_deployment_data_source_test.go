// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAwsWorkspacesDeploymentDataSourcePreCheck(t *testing.T) {

	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_GUID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_GUID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_ACCOUNT_GUID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_ACCOUNT_GUID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_IMAGE_GUID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_IMAGE_GUID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_CONNECTION_GUID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_CONNECTION_GUID must be set for acceptance tests")
	}
}

func TestAwsWorkspacesDeploymentDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_GUID")
	name := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_NAME")
	accountId := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_ACCOUNT_GUID")
	imageId := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_IMAGE_GUID")
	connectionId := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_DATA_SOURCE_CONNECTION_GUID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAwsWorkspacesAccountDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildAWSWorkspacesDeploymentDataSourceUsingId(t, aws_workspaces_deployment_test_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "id", id),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", name),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "account_id", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "image_id", imageId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "directory_connection_id", connectionId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildAWSWorkspacesDeploymentDataSourceUsingName(t, aws_workspaces_deployment_test_data_source_using_name, name, accountId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "id", id),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", name),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "account_id", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "image_id", imageId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "directory_connection_id", connectionId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesDeploymentDataSourceUsingId(t *testing.T, workspaceDeploymentDataSource string, id string) string {
	return fmt.Sprintf(workspaceDeploymentDataSource, id)
}

func BuildAWSWorkspacesDeploymentDataSourceUsingName(t *testing.T, workspaceDeploymentDataSource string, name string, accountId string) string {
	return fmt.Sprintf(workspaceDeploymentDataSource, name, accountId)
}

var (
	aws_workspaces_deployment_test_data_source_using_id = `
	data "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		id         = "%s"
	}
	`

	aws_workspaces_deployment_test_data_source_using_name = `
	data "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		name       = "%s"
		account_id = "%s"
	}
	`
)
