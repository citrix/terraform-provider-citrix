// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExampleValidateIPFilter() {
	// Used within a Schema method of a DataSource, Provider, or Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				Validators: []validator.String{
					ValidateIPFilter(),
					// Example Valid IP Filter Addresses
					// "*.*.*.*"
					// "12.0.0.*"
					// "12.0.*.*"
					// "12.*.*.*"
					// "12.0.0.0"
					// "12.0.0.1-12.0.0.70"
					// "12.0.0.1/24"
					// "2001:0db8:3c4d:0015:0:0:abcd:ef12"
					// "2001:0db8:3c4d:0015:0:0::"
					// "::3c4d:0015:0:0:abcd:ef12"
					// "2001:0db8:3c4d:0015:0:0:abcd:ef12/39"

					// Example Invalid IP Filter Addresses
					// "*.1.*.*"
					// "12.0.*.0"
					// "12.*.0.0"
					// "12.*.*.1"
					// "12.*.1.1"
					// "*.0.0.0"
					// "12.0.0.*/16"
					// "12.0.0.256"
					// "12.0.0.1/24-12.0.0.70"
					// "12.0.0.1-12.0.0.70/24"
					// "12.0.0.1/24-12.0.0.70/24"
					// "12.0.0.70-12.0.0.70"
					// "12.0.0.71-12.0.0.70"
					// "12.0.0.1/40"
					// "2001:0db8:3c4d:0015:0:0:abcd:ef12/asd"
					// "2001:0db8:3c4d:0015:0:0:"
					// "2001:0db8:3c4d:0015:0:0:::"
					// ":3c4d:0015:0:0:abcd:ef12"
					// ":::3c4d:0015:0:0:abcd:ef12"
					// "2001:0db8:3c4d:0015:0:0:abcd:ef12/40"
					// "2001:0db8:3c4d:0015:0:0:abcd:ef12/"
					// ""
				},
			},
		},
	}
}
