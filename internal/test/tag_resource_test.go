// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestTagResourcePreCheck validates the necessary env variable exist
// in the testing environment
func TestTagResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_TAG_RESOURCE_NAME"); v == "" {
		t.Fatal("TEST_TAG_RESOURCE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TAG_RESOURCE_DESCRIPTION"); v == "" {
		t.Fatal("TEST_TAG_RESOURCE_DESCRIPTION must be set for acceptance tests")
	}
}

func TestTagResource(t *testing.T) {
	tagName := os.Getenv("TEST_TAG_RESOURCE_NAME")
	tagDescription := os.Getenv("TEST_TAG_RESOURCE_DESCRIPTION")

	tagName_Updated := os.Getenv("TEST_TAG_RESOURCE_NAME") + "-updated"
	tagDescription_Updated := os.Getenv("TEST_TAG_RESOURCE_DESCRIPTION") + " description updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestTagResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: BuildTagResource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the name of the tag
					resource.TestCheckResourceAttr("citrix_tag.test_tag", "name", tagName),
					// Verify the description of tag
					resource.TestCheckResourceAttr("citrix_tag.test_tag", "description", tagDescription),
				),
			},
			// Import test
			{
				ResourceName:            "citrix_tag.test_tag",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update and Read test
			{
				Config: BuildTagResource_Updated(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the id of the tag
					resource.TestCheckResourceAttr("citrix_tag.test_tag", "name", tagName_Updated),
					// Verify the name of tag
					resource.TestCheckResourceAttr("citrix_tag.test_tag", "description", tagDescription_Updated),
				),
			},
		},
	})
}

func BuildTagResource(t *testing.T) string {
	tagName := os.Getenv("TEST_TAG_RESOURCE_NAME")
	tagDescription := os.Getenv("TEST_TAG_RESOURCE_DESCRIPTION")

	return fmt.Sprintf(tag_test_resource, tagName, tagDescription)
}

func BuildTagResource_Updated(t *testing.T) string {
	tagName := os.Getenv("TEST_TAG_RESOURCE_NAME") + "-updated"
	tagDescription := os.Getenv("TEST_TAG_RESOURCE_DESCRIPTION") + " description updated"

	return fmt.Sprintf(tag_test_resource, tagName, tagDescription)
}

var (
	tag_test_resource = `
	resource "citrix_tag" "test_tag" {
		name = "%s"
		description = "%s"
	}
	`
)
