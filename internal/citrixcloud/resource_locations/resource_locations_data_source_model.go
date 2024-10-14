// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	ccresourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceLocationsDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (ResourceLocationsDataSourceModel) GetSchema() schema.Schema {
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
		},
	}
}

func (r ResourceLocationsDataSourceModel) RefreshPropertyValues(ccResourceLocation *ccresourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel) ResourceLocationsDataSourceModel {
	// Overwrite resource location with refreshed state
	r.Id = types.StringValue(ccResourceLocation.GetId())
	r.Name = types.StringValue(ccResourceLocation.GetName())

	return r
}
