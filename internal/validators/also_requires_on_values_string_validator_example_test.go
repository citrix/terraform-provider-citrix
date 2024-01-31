// Copyright © 2023. Citrix Systems, Inc.

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExampleAlsoRequiresOnValues() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_attr": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					// Validate other_attr must be configured if this attribute is set to one of ["value1", "value2"].
					AlsoRequiresOnValues(
						[]string{
							"value1",
							"value2",
						},
						path.Expressions{
							path.MatchRoot("other_attr"),
						}...),
				},
			},
			"other_attr": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
