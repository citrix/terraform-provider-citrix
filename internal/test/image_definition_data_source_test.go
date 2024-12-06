// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestImageDefinitionDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, imageDefinitionDataSourceTestVariables)
}

func TestImageDefinitionDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestImageDefinitionDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Test Image Definition Data Source with Id
			{
				Config: BuildImageDefinitionDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "id", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_ID")),
					// Verify the name of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "name", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_NAME")),
					// Verify the description of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "description", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_DESCRIPTION")),
					// Verify the name of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "os_type", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_OS_TYPE")),
					// Verify the session_support of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "session_support", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_SESSION_SUPPORT")),
				),
			},
			// Test Image Definition Data Source with Name
			{
				Config: BuildImageDefinitionDataSourceWithName(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "id", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_ID")),
					// Verify the name of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "name", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_NAME")),
					// Verify the description of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "description", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_DESCRIPTION")),
					// Verify the name of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "os_type", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_OS_TYPE")),
					// Verify the session_support of the image definition data source
					resource.TestCheckResourceAttr("data.citrix_image_definition.test_image_definition", "session_support", os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_SESSION_SUPPORT")),
				),
			},
		},
	})
}

func BuildImageDefinitionDataSourceWithId(t *testing.T) string {
	imageDefinitionId := os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_ID")

	return fmt.Sprintf(imageDefinitionTestDataSourceById, imageDefinitionId)
}

func BuildImageDefinitionDataSourceWithName(t *testing.T) string {
	imageDefinitionName := os.Getenv("TEST_IMAGE_DEFINITION_DATA_SOURCE_NAME")

	return fmt.Sprintf(imageDefinitionTestDataSourceByName, imageDefinitionName)
}

var (
	imageDefinitionDataSourceTestVariables = []string{
		"TEST_IMAGE_DEFINITION_DATA_SOURCE_ID",
		"TEST_IMAGE_DEFINITION_DATA_SOURCE_NAME",
		"TEST_IMAGE_DEFINITION_DATA_SOURCE_DESCRIPTION",
		"TEST_IMAGE_DEFINITION_DATA_SOURCE_OS_TYPE",
		"TEST_IMAGE_DEFINITION_DATA_SOURCE_SESSION_SUPPORT",
	}

	imageDefinitionTestDataSourceById = `
data "citrix_image_definition" "test_image_definition" {
	id 			= "%s"
}
`

	imageDefinitionTestDataSourceByName = `
data "citrix_image_definition" "test_image_definition" {
	name 			= "%s"
}
`
)
