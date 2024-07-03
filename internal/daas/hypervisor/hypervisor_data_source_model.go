// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorDataSourceModel defines the Hypervisor data source implementation.
type HypervisorDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (HypervisorDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Read data of an existing hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
				Required:    true,
			},
		},
	}
}

func (r HypervisorDataSourceModel) RefreshPropertyValues(hypervisor *citrixorchestration.HypervisorDetailResponseModel) HypervisorDataSourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())

	return r
}
