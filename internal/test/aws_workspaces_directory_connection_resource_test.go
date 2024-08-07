// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAWSWorkspacesDirectoryConnectionPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1 must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2 must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED must be set for acceptance tests")
	}
}

func TestAWSWorkspacesDirectoryConnectionWithResourceLocationResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesAccountResourcePreCheck(t)
			TestAWSWorkspacesDirectoryConnectionPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Directory Connection resource with Resource Location
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "name", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "resource_location", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "directory", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1")),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "tenancy", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "security_group", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "default_ou", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "user_enabled_as_local_administrator", "false"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location",
				ImportState:             true,
				ImportStateIdFunc:       generateAwsDirectoryConnectionWithResourceLocationImportStateId,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resource_location", "zone"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Update testing for QCS AWS Workspaces Directory Connection resource with Resource Location
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation_Updated(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "name", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")+"-updated"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "resource_location", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "directory", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED")),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "tenancy", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "security_group", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "default_ou", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location", "user_enabled_as_local_administrator", "true"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t *testing.T) string {
	name := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")
	resourceLocation := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION")
	directory := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY")
	subnet1 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1")
	subnet2 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2")
	tenancy := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY")
	securityGroup := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP")
	defaultOU := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU")
	return fmt.Sprintf(testAWSWorkspacesDirectoryConnectionResource_withResourceLocation, name, resourceLocation, directory, subnet1, subnet2, tenancy, securityGroup, defaultOU, "false")
}

func BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation_Updated(t *testing.T) string {
	name := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME") + "-updated"
	resourceLocation := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_RESOURCE_LOCATION_UPDATED")
	directory := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED")
	subnet1 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED")
	subnet2 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED")
	tenancy := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED")
	securityGroup := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED")
	defaultOU := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED")
	return fmt.Sprintf(testAWSWorkspacesDirectoryConnectionResource_withResourceLocation, name, resourceLocation, directory, subnet1, subnet2, tenancy, securityGroup, defaultOU, "true")
}

func generateAwsDirectoryConnectionWithResourceLocationImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["account"], rawState["id"]), nil
}

func TestAWSWorkspacesDirectoryConnectionWithZoneIdResource(t *testing.T) {
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	isOnPremises := true
	if customerId != "" && customerId != "CitrixOnPremises" {
		// Tests being run in cloud env
		isOnPremises = false
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesAccountResourcePreCheck(t)
			TestAWSWorkspacesDirectoryConnectionPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Directory Connection resource with Zone
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithZoneId(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "name", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "zone", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "directory", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1")),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "tenancy", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "security_group", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "default_ou", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "user_enabled_as_local_administrator", "false"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id",
				ImportState:             true,
				ImportStateIdFunc:       generateAwsDirectoryConnectionWithZoneImportStateId,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resource_location", "zone"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},
			// Update testing for QCS AWS Workspaces Directory Connection resource with Zone
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithZoneId_Updated(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "name", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")+"-updated"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "zone", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "directory", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED")),
					resource.TestCheckTypeSetElemAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "subnets.*", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "tenancy", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "security_group", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "default_ou", os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id", "user_enabled_as_local_administrator", "true"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesDirectoryConnectionResourceWithZoneId(t *testing.T) string {
	name := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME")
	zoneId := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID")
	directory := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY")
	subnet1 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1")
	subnet2 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2")
	tenancy := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY")
	securityGroup := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP")
	defaultOU := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU")
	return fmt.Sprintf(testAWSWorkspacesDirectoryConnectionResource_withZoneId, name, zoneId, directory, subnet1, subnet2, tenancy, securityGroup, defaultOU, "false")
}

func BuildAWSWorkspacesDirectoryConnectionResourceWithZoneId_Updated(t *testing.T) string {
	name := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_NAME") + "-updated"
	zoneId := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_ZONE_ID_UPDATED")
	directory := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DIRECTORY_UPDATED")
	subnet1 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET1_UPDATED")
	subnet2 := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SUBNET2_UPDATED")
	tenancy := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_TENANCY_UPDATED")
	securityGroup := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_SECURITY_GROUP_UPDATED")
	defaultOU := os.Getenv("TEST_AWS_WORKSPACES_DIRECTORY_CONNECTION_DEFAULT_OU_UPDATED")
	return fmt.Sprintf(testAWSWorkspacesDirectoryConnectionResource_withZoneId, name, zoneId, directory, subnet1, subnet2, tenancy, securityGroup, defaultOU, "true")
}

func generateAwsDirectoryConnectionWithZoneImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_zone_id"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["account"], rawState["id"]), nil
}

var (
	testAWSWorkspacesDirectoryConnectionResource_withResourceLocation = `
	resource "citrix_quickcreate_aws_workspaces_directory_connection" "test_qcs_aws_dir_conn_with_resource_location" {
		name                                = "%s"
		account                             = citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		resource_location                   = "%s"
		directory                           = "%s"
		subnets                             = ["%s", "%s"]
		tenancy                             = "%s"
		security_group                      = "%s"
		default_ou                          = "%s"
		user_enabled_as_local_administrator = %s
	}
	`

	testAWSWorkspacesDirectoryConnectionResource_withZoneId = `
	resource "citrix_quickcreate_aws_workspaces_directory_connection" "test_qcs_aws_dir_conn_with_zone_id" {
		name                                = "%s"
		account                             = citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		zone                                = "%s"
		directory                           = "%s"
		subnets                             = ["%s", "%s"]
		tenancy                             = "%s"
		security_group                      = "%s"
		default_ou                          = "%s"
		user_enabled_as_local_administrator = %s
	}
	`
)
