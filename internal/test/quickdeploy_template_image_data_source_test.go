// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestQuickDeployTemplateImageDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestQuickDeployTemplateImageDataSourcePreCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	if v := os.Getenv("TEST_QUICKDEPLOY_IMAGE_ID"); v == "" {
		t.Fatal("TEST_QUICKDEPLOY_IMAGE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_QUICKDEPLOY_IMAGE_NAME"); v == "" {
		t.Fatal("TEST_QUICKDEPLOY_IMAGE_NAME must be set for acceptance tests")
	}
}

func TestQuickDeployTemplateImageDataSource(t *testing.T) {
	id := os.Getenv("TEST_QUICKDEPLOY_IMAGE_ID")
	name := os.Getenv("TEST_QUICKDEPLOY_IMAGE_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestQuickDeployTemplateImageDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Name
			{
				Config: BuildQuickDeployTemplateImageDataSource(t, quickdeploy_template_image_test_data_source_using_name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the quick deploy template image
					resource.TestCheckResourceAttr("data.citrix_quickdeploy_template_image.test_image", "id", id),
					// Verify the Name of the quick deploy template image
					resource.TestCheckResourceAttr("data.citrix_quickdeploy_template_image.test_image", "name", name),
				),
			},
		},
	})
}

func BuildQuickDeployTemplateImageDataSource(t *testing.T, quickDeployTemplateImageDataSource, templateImageName string) string {
	return fmt.Sprintf(quickDeployTemplateImageDataSource, templateImageName)
}

var (
	quickdeploy_template_image_test_data_source_using_name = `
	data citrix_quickdeploy_template_image test_image {
		name = "%s"
	}
	`
)
