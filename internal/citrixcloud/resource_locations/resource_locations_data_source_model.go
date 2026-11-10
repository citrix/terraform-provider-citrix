// Copyright © 2026. Citrix Systems, Inc.

package resource_locations

import (
	ccresourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResourceLocationDataSourceModel maps the data source schema data.
type ResourceLocationDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	InternalOnly types.Bool   `tfsdk:"internal_only"`
	TimeZone     types.String `tfsdk:"time_zone"`
}

func (ResourceLocationDataSourceModel) GetSchema() schema.Schema {
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

func (r ResourceLocationDataSourceModel) RefreshPropertyValues(ccResourceLocation *ccresourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel) ResourceLocationDataSourceModel {
	r.Id = types.StringValue(ccResourceLocation.GetId())
	r.Name = types.StringValue(ccResourceLocation.GetName())
	r.InternalOnly = types.BoolValue(ccResourceLocation.GetInternalOnly())
	r.TimeZone = types.StringValue(ccResourceLocation.GetTimeZone())

	return r
}
