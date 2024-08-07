// Copyright © 2024. Citrix Systems, Inc.

package zone

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ZoneDataSourceModel defines the Zone data source implementation.
type ZoneDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (ZoneDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the zone.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the zone.",
				Required:    true,
			},
		},
	}
}

func (r ZoneDataSourceModel) RefreshPropertyValues(zone *citrixorchestration.ZoneDetailResponseModel) ZoneDataSourceModel {
	r.Id = types.StringValue(zone.GetId())
	r.Name = types.StringValue(zone.GetName())

	return r
}
