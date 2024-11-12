// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (ResourceLocationModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Read data of an existing resource location.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the resource location.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource location.",
				Required:    true,
			},
			"internal_only": schema.BoolAttribute{
				Description: "Flag to determine if the resource location can only be used internally. Defaults to `false`.",
				Computed:    true,
			},
			"time_zone": schema.StringAttribute{
				Description: "Timezone associated with the resource location. Please refer to the `Timezone` column in the following [table](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/default-time-zones?view=windows-11#time-zones) for allowed values.",
				Computed:    true,
			},
		},
	}
}
