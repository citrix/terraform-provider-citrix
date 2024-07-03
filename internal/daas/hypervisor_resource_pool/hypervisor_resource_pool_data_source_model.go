// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourcePoolDataSourceModel defines the Hypervisor Resource Pool data source implementation.
type HypervisorResourcePoolDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	HypervisorName types.String `tfsdk:"hypervisor_name"`
	Networks       types.List   `tfsdk:"networks"` // List[string]
}

func (HypervisorResourcePoolDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Read data of an existing hypervisor resource pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor resource pool.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor resource pool.",
				Required:    true,
			},
			"hypervisor_name": schema.StringAttribute{
				Description: "Name of the hypervisor to which the resource pool belongs.",
				Required:    true,
			},
			"networks": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Networks available in the hypervisor resource pool.",
				Computed:    true,
			},
		},
	}
}

func (r HypervisorResourcePoolDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) HypervisorResourcePoolDataSourceModel {
	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.HypervisorName = types.StringValue(hypervisorConnection.GetName())

	var res []string
	if resourcePool.GetConnectionType() == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		for _, model := range resourcePool.GetSubnets() {
			res = append(res, model.GetName())
		}
	} else {
		for _, model := range resourcePool.GetNetworks() {
			res = append(res, model.GetName())
		}
	}

	r.Networks = util.RefreshListValues(ctx, diagnostics, r.Networks, res)

	return r
}
