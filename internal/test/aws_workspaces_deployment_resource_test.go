// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAWSWorkspacesDeploymentResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1 must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER2"); v == "" {
		t.Fatal("TEST_AWS_WORKSPACES_DEPLOYMENT_USER2 must be set for acceptance tests")
	}
}

func TestAWSWorkspacesUserCoupledDeploymentResource(t *testing.T) {
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
			TestAWSWorkspacesImageResourcePreCheck(t)
			TestAWSWorkspacesDirectoryConnectionPreCheck(t)
			TestAWSWorkspacesDeploymentResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Deployment resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResourceUpdated(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t),
					BuildAWSWorkspacesUserCoupledDeploymentResource(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Workspaces Deployment Tests
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "performance", "STANDARD"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "root_volume_size", "80"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_volume_size", "50"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encrypted", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "running_mode", "MANUAL"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "scale_settings.disconnect_session_idle_timeout", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_decoupled_workspaces", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.#", "2"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.username", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.maintenance_mode", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.1.username", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER2")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.1.maintenance_mode", "true"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scale_settings", "workspaces"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},

			// Update testing for QCS AWS Workspaces Deployment resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResourceUpdated(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t),
					BuildAWSWorkspacesUserCoupledDeploymentResourceUpdated(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")+"-updated"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "performance", "STANDARD"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "root_volume_size", "80"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_volume_size", "100"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encrypted", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "running_mode", "ALWAYS_ON"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_decoupled_workspaces", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.#", "1"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.username", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.maintenance_mode", "false"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesUserCoupledDeploymentResource(t *testing.T) string {
	deploymentName := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")
	sesionIdleTimeout := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT")
	deploymentUser1 := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1")
	deploymentUser2 := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER2")

	return fmt.Sprintf(testAWSWorkspacesUserCoupledDeploymentResource, deploymentName, sesionIdleTimeout, deploymentUser1, deploymentUser2)
}

func BuildAWSWorkspacesUserCoupledDeploymentResourceUpdated(t *testing.T) string {
	deploymentName := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")
	deploymentUser1 := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_USER1")
	return fmt.Sprintf(testAWSWorkspacesUserCoupledDeploymentResource_updated, deploymentName, deploymentUser1)
}

func TestAWSWorkspacesUserDecoupledDeploymentResource(t *testing.T) {
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
			TestAWSWorkspacesImageResourcePreCheck(t)
			TestAWSWorkspacesDirectoryConnectionPreCheck(t)
			TestAWSWorkspacesDeploymentResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Deployment resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResourceUpdated(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t),
					BuildAWSWorkspacesUserDecoupledDeploymentResource(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Workspaces Deployment Tests
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "performance", "STANDARD"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "root_volume_size", "80"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_volume_size", "50"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encrypted", "true"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encryption_key", "alias/aws/workspaces"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "running_mode", "MANUAL"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "scale_settings.disconnect_session_idle_timeout", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT")),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_decoupled_workspaces", "true"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.#", "2"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.maintenance_mode", "false"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.1.maintenance_mode", "true"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scale_settings", "workspaces"},
				SkipFunc:                skipForOnPrem(isOnPremises),
			},

			// Update testing for QCS AWS Workspaces Deployment resource
			{
				Config: composeTestResourceTf(
					BuildAWSWorkspacesAccountResourceWithARN(t),
					BuildAWSWorkspacesImageResourceUpdated(t),
					BuildAWSWorkspacesDirectoryConnectionResourceWithResourceLocation(t),
					BuildAWSWorkspacesUserDecoupledDeploymentResourceUpdated(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "name", os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")+"-updated"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "performance", "STANDARD"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "root_volume_size", "80"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_volume_size", "100"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encrypted", "true"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "volumes_encryption_key", "alias/aws/workspaces"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "running_mode", "ALWAYS_ON"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "user_decoupled_workspaces", "true"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.#", "1"),
					resource.TestCheckResourceAttr("citrix_quickcreate_aws_workspaces_deployment.test_aws_workspaces_deployment", "workspaces.0.maintenance_mode", "false"),
				),
				SkipFunc: skipForOnPrem(isOnPremises),
			},
		},
	})
}

func BuildAWSWorkspacesUserDecoupledDeploymentResource(t *testing.T) string {
	deploymentName := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")
	sesionIdleTimeout := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_SESSION_IDLE_TIMEOUT")

	return fmt.Sprintf(testAWSWorkspacesUserDecoupledDeploymentResource, deploymentName, sesionIdleTimeout)
}

func BuildAWSWorkspacesUserDecoupledDeploymentResourceUpdated(t *testing.T) string {
	deploymentName := os.Getenv("TEST_AWS_WORKSPACES_DEPLOYMENT_NAME")

	return fmt.Sprintf(testAWSWorkspacesUserDecoupledDeploymentResource_updated, deploymentName)
}

var (
	testAWSWorkspacesUserCoupledDeploymentResource = `
	resource "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		name 			  		= "%s"
		account_id 		  		= citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		directory_connection_id	= citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location.id
		image_id 		  		= citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image.id
		performance 	  		= "STANDARD"
		root_volume_size  		= "80"
		user_volume_size  		= "50"
		volumes_encrypted 		= false

		running_mode   = "MANUAL"
		scale_settings = {
			disconnect_session_idle_timeout = %s
			shutdown_disconnect_timeout = 15
			shutdown_log_off_timeout = 15
			buffer_capacity_size_percentage = 0
		}

		user_decoupled_workspaces = false
		workspaces = [
			{
				username = "%s"
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = false
			},
			{
				username = "%s"
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = true
			},
		]
	}
	`
	testAWSWorkspacesUserCoupledDeploymentResource_updated = `
	resource "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		account_id 		  		= citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location.id
		image_id 		  		= citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image.id
		name 			  		= "%s-updated"
		performance 	  		= "STANDARD"
		root_volume_size  		= "80"
		user_volume_size  		= "100"
		volumes_encrypted 		= false

		running_mode   = "ALWAYS_ON"

		user_decoupled_workspaces = false
		workspaces = [
			{
				username = "%s"
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = false
			},
		]
	}
	`

	testAWSWorkspacesUserDecoupledDeploymentResource = `
	resource "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		name 			  		= "%s"
		account_id 		  		= citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location.id
		image_id 		  		= citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image.id
		performance 	  		= "STANDARD"
		root_volume_size  		= "80"
		user_volume_size  		= "50"
		volumes_encrypted 		= true
		volumes_encryption_key 	= "alias/aws/workspaces"

		running_mode   = "MANUAL"
		scale_settings = {
			disconnect_session_idle_timeout = %s
			shutdown_disconnect_timeout = 15
			shutdown_log_off_timeout = 15
			buffer_capacity_size_percentage = 0
		}

		user_decoupled_workspaces = true
		workspaces = [
			{
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = false
			},
			{
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = true
			},
		]
	}
	`
	testAWSWorkspacesUserDecoupledDeploymentResource_updated = `
	resource "citrix_quickcreate_aws_workspaces_deployment" "test_aws_workspaces_deployment" {
		account_id 		  		= citrix_quickcreate_aws_workspaces_account.test_aws_workspaces_account_role_arn.id
		directory_connection_id= citrix_quickcreate_aws_workspaces_directory_connection.test_qcs_aws_dir_conn_with_resource_location.id
		image_id 		  		= citrix_quickcreate_aws_workspaces_image.test_aws_workspaces_image.id
		name 			  		= "%s-updated"
		performance 	  		= "STANDARD"
		root_volume_size  		= "80"
		user_volume_size  		= "100"
		volumes_encrypted 		= true
		volumes_encryption_key 	= "alias/aws/workspaces"

		running_mode   = "ALWAYS_ON"

		user_decoupled_workspaces = true
		workspaces = [
			{
				root_volume_size = 80
				user_volume_size = 50
				maintenance_mode = false
			},
		]
	}
	`
)
