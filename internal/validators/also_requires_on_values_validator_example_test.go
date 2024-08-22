// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExampleAlsoRequiresOnStringValues() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_string_attr": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					// Validate other_attr must be configured if this attribute is set to one of ["value1", "value2"].
					AlsoRequiresOnStringValues(
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

func ExampleAlsoRequiresOnBoolValues() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_bool_attr": schema.BoolAttribute{
				Optional: true,
				Validators: []validator.Bool{
					// Validate other_attr must be configured if this attribute is set to one of [true].
					AlsoRequiresOnBoolValues(
						[]bool{
							true,
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
