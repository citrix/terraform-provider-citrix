// Copyright © 2024. Citrix Systems, Inc.
package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAWSWorkspacesCloudFormationTemplateDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: BuildAWSWorkspacesCloudFormationTemplateDataSource(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrix_quickcreate_aws_workspaces_cloudformation_template.test_cloudformation_template", "content"),
				),
			},
		},
	})
}

func BuildAWSWorkspacesCloudFormationTemplateDataSource(t *testing.T) string {
	return cloudFormationTemplateTestDataSource
}

var (
	cloudFormationTemplateTestDataSource = `
	data "citrix_quickcreate_aws_workspaces_cloudformation_template" "test_cloudformation_template" { }
	`
)
