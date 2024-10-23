// Copyright © 2024. Citrix Systems, Inc.

package wem_site

import (
	"context"
	"strconv"

	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WemSiteDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r WemSiteDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, wemSite *citrixwemservice.SiteModel) WemSiteDataSourceModel {

	r.Id = types.StringValue(strconv.FormatInt(wemSite.GetId(), 10))
	r.Name = types.StringValue(wemSite.GetName())
	r.Description = types.StringValue(wemSite.GetDescription())

	return r
}

func GetWemSiteDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "WEM --- Data source to get details regarding a specific configuration set.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the WEM configuration set.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the configuration set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the configuration set.",
				Computed:    true,
			},
		},
	}
}
