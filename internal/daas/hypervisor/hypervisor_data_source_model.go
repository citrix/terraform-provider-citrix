// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorDataSourceModel defines the Hypervisor data source implementation.
type HypervisorDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Tenants types.Set    `tfsdk:"tenants"` // Set[string]
}

func (HypervisorDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
				Required:    true,
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the hypervisor connection.",
				Computed:    true,
			},
		},
	}
}

func (r HypervisorDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) HypervisorDataSourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, hypervisor.GetTenants())

	return r
}
