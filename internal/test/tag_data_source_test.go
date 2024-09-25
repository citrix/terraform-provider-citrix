// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestTagDataSourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestTagDataSourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_TAG_DATA_SOURCE_ID"); v == "" {
		t.Fatal("TEST_TAG_DATA_SOURCE_ID must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TAG_DATA_SOURCE_NAME"); v == "" {
		t.Fatal("TEST_TAG_DATA_SOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TAG_DATA_SOURCE_DESCRIPTION"); v == "" {
		t.Fatal("TEST_TAG_DATA_SOURCE_DESCRIPTION must be set for acceptance tests")
	}
}

func TestTagDataSource(t *testing.T) {
	tagId := os.Getenv("TEST_TAG_DATA_SOURCE_ID")
	tagName := os.Getenv("TEST_TAG_DATA_SOURCE_NAME")
	tagDescription := os.Getenv("TEST_TAG_DATA_SOURCE_DESCRIPTION")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestTagDataSourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing using Tag ID
			{
				Config: BuildTagDataSource(t, tag_test_data_source_by_id, tagId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "id", tagId),
					// Verify the name of tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "name", tagName),
					// Verify the description of tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "description", tagDescription),
				),
			},
			// Read testing using Tag Name
			{
				Config: BuildTagDataSource(t, tag_test_data_source_by_name, tagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "id", tagId),
					// Verify the name of tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "name", tagName),
					// Verify the description of tag
					resource.TestCheckResourceAttr("data.citrix_tag.test_tag", "description", tagDescription),
				),
			},
		},
	})
}

func BuildTagDataSource(t *testing.T, tagDataSource string, tagDataSourceAttribute string) string {
	return fmt.Sprintf(tagDataSource, tagDataSourceAttribute)
}

var (
	tag_test_data_source_by_id = `
	data "citrix_tag" "test_tag" {
		id = "%s"
	}
	`

	tag_test_data_source_by_name = `
	data "citrix_tag" "test_tag" {
		name = "%s"
	}
	`
)
