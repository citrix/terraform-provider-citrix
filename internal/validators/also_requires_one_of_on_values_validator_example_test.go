// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExampleAlsoRequiresOneOfOnStringValues() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_string_attr": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					// Validate exactly one of other_attr1 and other_attr2 must be configured if this attribute is set to one of ["value1", "value2"].
					AlsoRequiresOneOfOnStringValues(
						[]string{
							"value1",
							"value2",
						},
						path.Expressions{
							path.MatchRoot("other_attr1"),
							path.MatchRoot("other_attr2"),
						}...),
				},
			},
			"other_attr1": schema.StringAttribute{
				Optional: true,
			},
			"other_attr2": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func ExampleAlsoRequiresOneOfOnBoolValues() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_bool_attr": schema.BoolAttribute{
				Optional: true,
				Validators: []validator.Bool{
					// Validate exactly one of other_attr1 and other_attr2 must be configured if this attribute is set to one of [true].
					AlsoRequiresOneOfOnBoolValues(
						[]bool{
							true,
						},
						path.Expressions{
							path.MatchRoot("other_attr1"),
							path.MatchRoot("other_attr2"),
						}...),
				},
			},
			"other_attr1": schema.StringAttribute{
				Optional: true,
			},
			"other_attr2": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
