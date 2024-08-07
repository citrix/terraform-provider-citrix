// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAWSWorkspacesDirectoryConnectionDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_EXPECTED_TENANCY"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_EXPECTED_TENANCY must be set for acceptance tests")
	}
}

func TestAWSWorkspacesDirectoryConnectionDataSource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	directoryConnectionId := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_ID")
	directoryConnectionName := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DATA_SOURCE_NAME")
	accountId := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ID")
	tenancy := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_EXPECTED_TENANCY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesDirectoryConnectionDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Directory Connection ID and Account Id
			{
				Config: BuildAWSWorkspacesDirectoryConnectionDataSource(t, aws_workspaces_directory_connection_test_data_source_with_id, directoryConnectionId, accountId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "id", directoryConnectionId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "name", directoryConnectionName),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "subnets.#", "2"),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "account", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "tenancy", tenancy),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
			// Read testing using Directory Connection Name and Account Id
			{
				Config: BuildAWSWorkspacesDirectoryConnectionDataSource(t, aws_workspaces_directory_connection_test_data_source_with_name, directoryConnectionName, accountId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "id", directoryConnectionId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "name", directoryConnectionName),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "subnets.#", "2"),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "account", accountId),
					resource.TestCheckResourceAttr("data.citrix_quickcreate_aws_workspaces_directory_connection.test_aws_workspaces_directory_connection", "tenancy", tenancy),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesDirectoryConnectionDataSource(t *testing.T, directoryConnectionDataSource string, id string, accountId string) string {
	return fmt.Sprintf(directoryConnectionDataSource, id, accountId)
}

var (
	aws_workspaces_directory_connection_test_data_source_with_id = `
	data "citrix_quickcreate_aws_workspaces_directory_connection" "test_aws_workspaces_directory_connection" {
		id      = "%s"
		account = "%s"
	}
	`

	aws_workspaces_directory_connection_test_data_source_with_name = `
	data "citrix_quickcreate_aws_workspaces_directory_connection" "test_aws_workspaces_directory_connection" {
		name    = "%s"
		account = "%s"
	}
	`
)
