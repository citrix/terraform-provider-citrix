// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAwsWorkspacesAccountDataSourcePreCheck(t *testing.T) {

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_GUID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_GUID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_AWS_ACCOUNT_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_AWS_ACCOUNT_ID must be set for acceptance tests")
	}
}

func TestAwsWorkspacesAccountDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	id := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_GUID")
	name := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_NAME")
	awsAccountId := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_DATA_SOURCE_AWS_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAwsWorkspacesAccountDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: BuildAWSWorkspacesAccountDataSource(t, aws_workspaces_account_test_data_source_using_id, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "id", id),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "name", name),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "aws_account", awsAccountId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			{
				Config: BuildAWSWorkspacesAccountDataSource(t, aws_workspaces_account_test_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "id", id),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "name", name),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account", "aws_account", awsAccountId),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesAccountDataSource(t *testing.T, workspaceAccountDataSource string, idOrName string) string {
	return fmt.Sprintf(workspaceAccountDataSource, idOrName)
}

var (
	aws_workspaces_account_test_data_source_using_id = `
	data "citrix_quickcreate_aws_workspaces_account" "test_aws_workspaces_account" {
		id         = "%s"
	}
	`

	aws_workspaces_account_test_data_source_using_name = `
	data "citrix_quickcreate_aws_workspaces_account" "test_aws_workspaces_account" {
		name       = "%s"
	}
	`
)
