// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestImageVersionDataSourcePreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, imageVersionDataSourceTestVariables)
}

func TestImageVersionDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestImageVersionDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Test Image Version Data Source with Id
			{
				Config: BuildImageVersionDataSourceWithId(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "id", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_ID")),
					// Verify the version number of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "version_number", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_VERSION_NUMBER")),
					// Verify the image definition id of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "image_definition", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_IMAGE_DEFINITION_ID")),
					// Verify the session support of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "session_support", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_SESSION_SUPPORT")),
					// Verify the os type of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "os_type", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_OS_TYPE")),
				),
			},
			// Test Image Version Data Source with Version Number
			{
				Config: BuildImageVersionDataSourceWithVersionNumber(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "id", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_ID")),
					// Verify the version number of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "version_number", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_VERSION_NUMBER")),
					// Verify the image definition id of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "image_definition", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_IMAGE_DEFINITION_ID")),
					// Verify the session support of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "session_support", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_SESSION_SUPPORT")),
					// Verify the os type of the image version data source
					resource.TestCheckResourceAttr("data.citrix_image_version.test_image_version", "os_type", os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_OS_TYPE")),
				),
			},
		},
	})
}

func BuildImageVersionDataSourceWithId(t *testing.T) string {
	imageVersionId := os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_ID")
	imageDefinitionId := os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_IMAGE_DEFINITION_ID")

	return fmt.Sprintf(imageVersionTestDataSourceById, imageVersionId, imageDefinitionId)
}

func BuildImageVersionDataSourceWithVersionNumber(t *testing.T) string {
	imageVersionNumber := os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_VERSION_NUMBER")
	imageDefinitionId := os.Getenv("TEST_IMAGE_VERSION_DATA_SOURCE_IMAGE_DEFINITION_ID")

	return fmt.Sprintf(imageVersionTestDataSourceByNumber, imageVersionNumber, imageDefinitionId)
}

var (
	imageVersionDataSourceTestVariables = []string{
		"TEST_IMAGE_VERSION_DATA_SOURCE_ID",
		"TEST_IMAGE_VERSION_DATA_SOURCE_VERSION_NUMBER",
		"TEST_IMAGE_VERSION_DATA_SOURCE_IMAGE_DEFINITION_ID",
		"TEST_IMAGE_VERSION_DATA_SOURCE_SESSION_SUPPORT",
		"TEST_IMAGE_VERSION_DATA_SOURCE_OS_TYPE",
	}

	imageVersionTestDataSourceById = `
data "citrix_image_version" "test_image_version" {
	id 			     = "%s"
	image_definition = "%s"
}
`

	imageVersionTestDataSourceByNumber = `
data "citrix_image_version" "test_image_version" {
	version_number 	 = %s
	image_definition = "%s"
}
`
)
