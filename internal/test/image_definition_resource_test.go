// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAzureImageDefinitionResourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, imageDefinitionTestVariables)
}

func AzureImageDefinitionResourceHelper(t *testing.T, pre121 bool) {
	var imageDefinitionResource string
	var imageDefinitionResourceUpdated string
	var importStateVerifyIgnore []string
	if pre121 {
		imageDefinitionResource = BuildImageDefinitionTestResourcePre121(t)
		imageDefinitionResourceUpdated = BuildImageDefinitionUpdatedTestResourcePre121(t)
		importStateVerifyIgnore = []string{"hypervisor"}
	} else {
		imageDefinitionResource = BuildAzureImageDefinitionTestResource(t)
		imageDefinitionResourceUpdated = BuildAzureImageDefinitionUpdatedTestResource(t)
		importStateVerifyIgnore = []string{}
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAzureImageDefinitionResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: imageDefinitionResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "name", os.Getenv("TEST_IMAGE_DEFINITION_NAME")),
					// Verify the description of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "description", os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION")),
					// Verify the os_type of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "os_type", os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE")),
					// Verify the session_support of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "session_support", os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT")),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_image_definition.test_image_definition",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			// Update and Read testing
			{
				Config: imageDefinitionResourceUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "name", os.Getenv("TEST_IMAGE_DEFINITION_NAME_UPDATED")),
					// Verify the description of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "description", os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION_UPDATED")),
					// Verify the os_type of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "os_type", os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE_UPDATED")),
					// Verify the session_support of the image definition
					resource.TestCheckResourceAttr("citrix_image_definition.test_image_definition", "session_support", os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT_UPDATED")),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAzureImageDefinitionResource(t *testing.T) {
	AzureImageDefinitionResourceHelper(t, false)
}

func TestAzureImageDefinitionResourcePre121(t *testing.T) {
	AzureImageDefinitionResourceHelper(t, true)
}

func BuildAzureImageDefinitionTestResource(t *testing.T) string {
	imageDefinitionName := os.Getenv("TEST_IMAGE_DEFINITION_NAME")
	imageDefinitionDescription := os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION")
	imageDefinitionOsType := os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE")
	imageDefinitionSessionSupport := os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT")
	imageDefinitionHypervisorId := os.Getenv("TEST_IMAGE_DEFINITION_HYPERVISOR_ID")
	imageDefinitionResourceGroup := os.Getenv("TEST_IMAGE_DEFINITION_RESOURCE_GROUP")

	return fmt.Sprintf(imageDefinitionTestResource, imageDefinitionName, imageDefinitionDescription, imageDefinitionOsType, imageDefinitionSessionSupport, imageDefinitionHypervisorId, imageDefinitionResourceGroup)
}

func BuildAzureImageDefinitionUpdatedTestResource(t *testing.T) string {
	imageDefinitionName := os.Getenv("TEST_IMAGE_DEFINITION_NAME_UPDATED")
	imageDefinitionDescription := os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION_UPDATED")
	imageDefinitionOsType := os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE_UPDATED")
	imageDefinitionSessionSupport := os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT_UPDATED")
	imageDefinitionHypervisorId := os.Getenv("TEST_IMAGE_DEFINITION_HYPERVISOR_ID_UPDATED")
	imageDefinitionResourceGroup := os.Getenv("TEST_IMAGE_DEFINITION_RESOURCE_GROUP_UPDATED")

	return fmt.Sprintf(imageDefinitionTestResource, imageDefinitionName, imageDefinitionDescription, imageDefinitionOsType, imageDefinitionSessionSupport, imageDefinitionHypervisorId, imageDefinitionResourceGroup)
}

func BuildImageDefinitionTestResourcePre121(t *testing.T) string {
	imageDefinitionName := os.Getenv("TEST_IMAGE_DEFINITION_NAME")
	imageDefinitionDescription := os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION")
	imageDefinitionOsType := os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE")
	imageDefinitionSessionSupport := os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT")
	imageDefinitionHypervisorId := os.Getenv("TEST_IMAGE_DEFINITION_HYPERVISOR_ID")

	return fmt.Sprintf(imageDefinitionTestResourcePre121, imageDefinitionName, imageDefinitionDescription, imageDefinitionOsType, imageDefinitionSessionSupport, imageDefinitionHypervisorId)
}

func BuildImageDefinitionUpdatedTestResourcePre121(t *testing.T) string {
	imageDefinitionName := os.Getenv("TEST_IMAGE_DEFINITION_NAME_UPDATED")
	imageDefinitionDescription := os.Getenv("TEST_IMAGE_DEFINITION_DESCRIPTION_UPDATED")
	imageDefinitionOsType := os.Getenv("TEST_IMAGE_DEFINITION_OS_TYPE_UPDATED")
	imageDefinitionSessionSupport := os.Getenv("TEST_IMAGE_DEFINITION_SESSION_SUPPORT_UPDATED")
	imageDefinitionHypervisorId := os.Getenv("TEST_IMAGE_DEFINITION_HYPERVISOR_ID_UPDATED")

	return fmt.Sprintf(imageDefinitionTestResourcePre121, imageDefinitionName, imageDefinitionDescription, imageDefinitionOsType, imageDefinitionSessionSupport, imageDefinitionHypervisorId)
}

var (
	imageDefinitionTestVariables = []string{
		"TEST_IMAGE_DEFINITION_NAME",
		"TEST_IMAGE_DEFINITION_NAME_UPDATED",
		"TEST_IMAGE_DEFINITION_DESCRIPTION",
		"TEST_IMAGE_DEFINITION_DESCRIPTION_UPDATED",
		"TEST_IMAGE_DEFINITION_OS_TYPE",
		"TEST_IMAGE_DEFINITION_OS_TYPE_UPDATED",
		"TEST_IMAGE_DEFINITION_SESSION_SUPPORT",
		"TEST_IMAGE_DEFINITION_SESSION_SUPPORT_UPDATED",
		"TEST_IMAGE_DEFINITION_HYPERVISOR_ID",
		"TEST_IMAGE_DEFINITION_HYPERVISOR_ID_UPDATED",
		"TEST_IMAGE_DEFINITION_RESOURCE_GROUP",
		"TEST_IMAGE_DEFINITION_RESOURCE_GROUP_UPDATED",
	}

	imageDefinitionTestResource = `
resource "citrix_image_definition" "test_image_definition" {
	name 			= "%s"
	description 	= "%s"
	os_type 		= "%s"
	session_support = "%s"
	hypervisor      = "%s"
    azure_image_definition = {
		resource_group = "%s"
        use_image_gallery = false
    }
}
`

	imageDefinitionTestResourcePre121 = `
resource "citrix_image_definition" "test_image_definition" {
	name 			= "%s"
	description 	= "%s"
	os_type 		= "%s"
	session_support = "%s"
	hypervisor      = "%s"
}
`
)
