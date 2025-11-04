// Copyright © 2024. Citrix Systems, Inc.
package cma_vnet_peering

import (
	"context"

	catalogservice "github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CitrixManagedAzureOnPremNetworkConnectionDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (CitrixManagedAzureOnPremNetworkConnectionDataSourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - Citrix Managed Azure --- Data Source of an Citrix Managed Azure VNet Peering.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the VNet Peering.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the VNet Peering.",
				Required:    true,
			},
		},
	}
}

func (CitrixManagedAzureOnPremNetworkConnectionDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return CitrixManagedAzureOnPremNetworkConnectionDataSourceModel{}.GetDataSourceSchema().Attributes
}

func (r CitrixManagedAzureOnPremNetworkConnectionDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, onPremConnection *catalogservice.OnPremConnectionModel) CitrixManagedAzureOnPremNetworkConnectionDataSourceModel {
	r.Id = types.StringValue(onPremConnection.GetId())
	r.Name = types.StringValue(onPremConnection.GetName())

	return r
}
