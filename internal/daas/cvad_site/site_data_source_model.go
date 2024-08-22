// Copyright © 2024. Citrix Systems, Inc.
package cvad_site

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SiteDataSourceModel struct {
	SiteId                  types.String `tfsdk:"site_id"`
	CustomerId              types.String `tfsdk:"customer_id"`
	OrchestrationApiVersion types.Int64  `tfsdk:"orchestration_api_version"`
	ProductVersion          types.String `tfsdk:"product_version"`
	IsCitrixServiceProvider types.Bool   `tfsdk:"is_citrix_service_provider"`
	IsOnPremisesSite        types.Bool   `tfsdk:"is_on_premises"`
}

func (SiteDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get site information.",

		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "ID of the site.",
				Computed:    true,
			},
			"customer_id": schema.StringAttribute{
				Description: "ID of the customer the site belongs to.",
				Computed:    true,
			},
			"orchestration_api_version": schema.Int64Attribute{
				Description: "Version of the Orchestration Service API.",
				Computed:    true,
			},
			"product_version": schema.StringAttribute{
				Description: "Version of the product.",
				Computed:    true,
			},
			"is_citrix_service_provider": schema.BoolAttribute{
				Description: "Indicates whether the site belongs to a Citrix Service Provider.",
				Computed:    true,
			},
			"is_on_premises": schema.BoolAttribute{
				Description: "Indicates whether the site is an On Premises site.",
				Computed:    true,
			},
		},
	}
}

func (r SiteDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, clientConfig *citrixdaasclient.ClientConfiguration, isOnPrem bool) SiteDataSourceModel {
	r.SiteId = types.StringValue(clientConfig.SiteId)
	r.CustomerId = types.StringValue(clientConfig.CustomerId)
	r.OrchestrationApiVersion = types.Int64Value(int64(clientConfig.OrchestrationApiVersion))
	r.ProductVersion = types.StringValue(clientConfig.ProductVersion)
	r.IsCitrixServiceProvider = types.BoolValue(clientConfig.IsCspCustomer)
	r.IsOnPremisesSite = types.BoolValue(isOnPrem)

	return r
}
