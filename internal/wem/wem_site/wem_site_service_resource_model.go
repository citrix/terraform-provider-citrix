// Copyright © 2024. Citrix Systems, Inc.
package wem_site

import (
	"context"
	"strconv"

	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WemSiteResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (WemSiteResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "WEM --- Manages configuration sets within a WEM deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the configuration site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the configuration site. WEM Site Name should be unique.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the configuration site. Default value is empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""), // Default value is empty string
			},
		},
	}
}

func (WemSiteResourceModel) GetAttributes() map[string]schema.Attribute {
	return WemSiteResourceModel{}.GetSchema().Attributes
}

func (r WemSiteResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, wemSite *citrixwemservice.SiteModel) WemSiteResourceModel {
	r.Id = types.StringValue(strconv.FormatInt(wemSite.GetId(), 10))
	r.Name = types.StringValue(wemSite.GetName())
	r.Description = types.StringValue(wemSite.GetDescription())
	return r
}
