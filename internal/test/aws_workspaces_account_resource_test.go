// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAWSWorkspacesAccountResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME_UPDATED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_NAME_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_REGION must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ACCESS_KEY_ID"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_ACCESS_KEY_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_SECRET_ACCESS_KEY"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_ACCOUNT_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
}

func TestAWSWorkspacesAccountResourceWithARN(t *testing.T) {
	accountName := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME") + "-role-arn"
	accountNameUpdated := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME_UPDATED") + "-role-arn"
	region := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION")
	byolFeatureEnabled := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED")
	roleArn := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesAccountResourcePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing for QCS AWS Workspaces Account resource with Role ARN
			{
				Config: BuildAWSWorkspacesAccountResourceWithARN(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Account with Role ARN Tests
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "name", accountName),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_region", region),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_byol_feature_enabled", byolFeatureEnabled),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_role_arn", roleArn),
				),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"aws_role_arn", "aws_access_key_id", "aws_secret_access_key"},
			},

			// Update testing for QCS AWS Workspaces Account resource
			{
				Config: BuildAWSWorkspacesAccountResourceWithARNUpdated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "name", accountNameUpdated),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_region", region),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_byol_feature_enabled", byolFeatureEnabled),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn", "aws_role_arn", roleArn),
				),
			},
		},
	})
}

func TestAWSWorkspacesAccountResourceWithAccessKey(t *testing.T) {
	accountName := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME") + "-access-key"
	region := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION")
	byolFeatureEnabled := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED")
	accessKeyId := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_SECRET_ACCESS_KEY")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAWSWorkspacesAccountResourcePreCheck(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing for QCS AWS Workspaces Account resource with Role ARN
			{
				Config: BuildAWSWorkspacesAccountResourceWithAccessKey(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Account with Role ARN Tests
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id", "name", accountName),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id", "aws_region", region),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id", "aws_byol_feature_enabled", byolFeatureEnabled),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id", "aws_access_key_id", accessKeyId),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id", "aws_secret_access_key", secretAccessKey),
				),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_access_key_id",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"aws_role_arn", "aws_access_key_id", "aws_secret_access_key"},
			},
		},
	})
}

func BuildAWSWorkspacesAccountResourceWithARN(t *testing.T) string {
	accountName := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME") + "-role-arn"
	region := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION")
	byolFeatureEnabled := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED")
	roleArn := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN")

	return fmt.Sprintf(testAWSWorkspacesAccountResource_withARN, accountName, region, byolFeatureEnabled, roleArn)
}

func BuildAWSWorkspacesAccountResourceWithARNUpdated(t *testing.T) string {
	accountNameUpdated := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME_UPDATED") + "-role-arn"
	region := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION")
	byolFeatureEnabled := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED")
	roleArn := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ROLE_ARN")

	return fmt.Sprintf(testAWSWorkspacesAccountResource_withARN_updated, accountNameUpdated, region, byolFeatureEnabled, roleArn)
}

func BuildAWSWorkspacesAccountResourceWithAccessKey(t *testing.T) string {
	accountName := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_NAME") + "-access-key"
	region := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_REGION")
	byolFeatureEnabled := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_BYOL_FEATURE_ENABLED")
	accessKeyId := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("TEST_AWS_WORKSPACES_ACCOUNT_SECRET_ACCESS_KEY")

	return fmt.Sprintf(testAWSWorkspacesAccountResource_withAccessKey, accountName, region, byolFeatureEnabled, accessKeyId, secretAccessKey)
}

var (
	testAWSWorkspacesAccountResource_withARN = `
	resource "citrix_quickcreate_aws_workspaces_account" "test_aws_workspaces_account_role_arn" {
		name                        = "%s"
		aws_region                  = "%s"
		aws_byol_feature_enabled    = "%s"
		aws_role_arn                = "%s"
	}
	`
	testAWSWorkspacesAccountResource_withARN_updated = `
	resource "citrix_quickcreate_aws_workspaces_account" "test_aws_workspaces_account_role_arn" {
		name                        = "%s"
		aws_region                  = "%s"
		aws_byol_feature_enabled    = "%s"
		aws_role_arn                = "%s"
	}
	`
	testAWSWorkspacesAccountResource_withAccessKey = `
	resource "citrix_quickcreate_aws_workspaces_account" "test_aws_workspaces_account_access_key_id" {
		name                        = "%s"
		aws_region                  = "%s"
		aws_byol_feature_enabled    = "%s"
		aws_access_key_id           = "%s"
		aws_secret_access_key       = "%s"
	}
	`
)
